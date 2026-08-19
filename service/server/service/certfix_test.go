package service

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/certpin"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
)

// sharedTestDir holds a single temporary directory used by all certfix tests.
// The v2rayA database connection is cached for the process lifetime, so
// switching directories between tests would cause each test to operate on the
// database file created by the first test.
var sharedTestDir string

func TestMain(m *testing.M) {
	var err error
	sharedTestDir, err = os.MkdirTemp("", "certfix-test-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(sharedTestDir)

	_ = os.Setenv("V2RAYA_CONFIG", sharedTestDir)
	conf.SetConfig(conf.Params{Config: sharedTestDir})
	// Initialize the database connection once so every test uses the same path.
	_ = db.GetDB()

	code := m.Run()
	os.Exit(code)
}

func resetConfig(t *testing.T) {
	t.Helper()
	if _, err := db.GetDB().Exec("DELETE FROM servers"); err != nil {
		t.Fatalf("DELETE FROM servers: %v", err)
	}
	if _, err := db.GetDB().Exec("DELETE FROM subscriptions"); err != nil {
		t.Fatalf("DELETE FROM subscriptions: %v", err)
	}
	if err := configure.SetConfigure(configure.New()); err != nil {
		t.Fatalf("SetConfigure: %v", err)
	}
}

func addServer(t *testing.T, protocol, link string) {
	t.Helper()
	obj, err := serverObj.NewFromLink(protocol, link)
	if err != nil {
		t.Fatalf("NewFromLink: %v", err)
	}
	if err := configure.AppendServers([]*configure.ServerRaw{{ServerObj: obj}}); err != nil {
		t.Fatalf("AppendServers: %v", err)
	}
}

func TestDetectCandidates_Empty(t *testing.T) {
	got, err := DetectCandidates(nil)
	if err != nil {
		t.Fatalf("DetectCandidates(nil) error = %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected no candidates, got %d", len(got))
	}
}

func TestDetectCandidates_SkipsNonTLS(t *testing.T) {
	resetConfig(t)
	addServer(t, "ss", "ss://YWVzLTEyOC1nY206dGVzdA==@127.0.0.1:8080#test")

	got, err := DetectCandidates([]*configure.Which{{TYPE: configure.ServerType, ID: 1}})
	if err != nil {
		t.Fatalf("DetectCandidates error = %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected non-TLS server to be skipped, got %d candidates", len(got))
	}
}

func TestStartFix_RejectConcurrent(t *testing.T) {
	resetConfig(t)
	addServer(t, "trojan", "trojan://user@127.0.0.1:443?allowInsecure=1#slow")

	originalProbe := certFixProbe
	certFixProbe = func(ctx context.Context, target *certpin.Target) (*certpin.Result, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	defer func() { certFixProbe = originalProbe }()

	job, err := StartFix([]CertFixCandidate{{Which: configure.Which{TYPE: configure.ServerType, ID: 1}, Name: "slow"}})
	if err != nil {
		t.Fatalf("StartFix error = %v", err)
	}
	defer CancelJob(job.ID)

	_, err = StartFix([]CertFixCandidate{{Which: configure.Which{TYPE: configure.ServerType, ID: 1}, Name: "second"}})
	if err == nil {
		t.Fatal("expected concurrent StartFix to be rejected")
	}
}

func TestGetJob_NotFound(t *testing.T) {
	_, err := GetJob("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing job")
	}
}

func TestCancelJob(t *testing.T) {
	resetConfig(t)
	addServer(t, "trojan", "trojan://user@127.0.0.1:443?allowInsecure=1#cancel")

	originalProbe := certFixProbe
	certFixProbe = func(ctx context.Context, target *certpin.Target) (*certpin.Result, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	defer func() { certFixProbe = originalProbe }()

	job, err := StartFix([]CertFixCandidate{{Which: configure.Which{TYPE: configure.ServerType, ID: 1}, Name: "cancel-me"}})
	if err != nil {
		t.Fatalf("StartFix error = %v", err)
	}

	if err := CancelJob(job.ID); err != nil {
		t.Fatalf("CancelJob error = %v", err)
	}

	time.Sleep(100 * time.Millisecond)

	j, err := GetJob(job.ID)
	if err != nil {
		t.Fatalf("GetJob error = %v", err)
	}
	if j.Status != "cancelled" {
		t.Fatalf("expected cancelled status, got %q", j.Status)
	}
}

func TestJobStateTransitions(t *testing.T) {
	resetConfig(t)
	addServer(t, "trojan", "trojan://user@127.0.0.1:443?allowInsecure=1#test")

	originalProbe := certFixProbe
	certFixProbe = func(ctx context.Context, target *certpin.Target) (*certpin.Result, error) {
		return &certpin.Result{Trusted: true}, nil
	}
	defer func() { certFixProbe = originalProbe }()

	job, err := StartFix([]CertFixCandidate{{Which: configure.Which{TYPE: configure.ServerType, ID: 1}, Name: "test"}})
	if err != nil {
		t.Fatalf("StartFix error = %v", err)
	}

	deadline := time.Now().Add(5 * time.Second)
	var final *CertFixJob
	for time.Now().Before(deadline) {
		final, err = GetJob(job.ID)
		if err != nil {
			t.Fatalf("GetJob error = %v", err)
		}
		if final.Status == "completed" || final.Status == "completed_with_failures" {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	if final.Status != "completed" {
		t.Fatalf("expected completed, got %q", final.Status)
	}
	if final.Total != 1 || final.Processed != 1 || final.Succeeded != 1 {
		t.Fatalf("unexpected counters: total=%d processed=%d succeeded=%d", final.Total, final.Processed, final.Succeeded)
	}
}

func TestStartFix_ProbeError(t *testing.T) {
	resetConfig(t)
	addServer(t, "trojan", "trojan://user@127.0.0.1:443?allowInsecure=1#test")

	originalProbe := certFixProbe
	certFixProbe = func(ctx context.Context, target *certpin.Target) (*certpin.Result, error) {
		return nil, errors.New("probe failed")
	}
	defer func() { certFixProbe = originalProbe }()

	job, err := StartFix([]CertFixCandidate{{Which: configure.Which{TYPE: configure.ServerType, ID: 1}, Name: "fail"}})
	if err != nil {
		t.Fatalf("StartFix error = %v", err)
	}

	deadline := time.Now().Add(5 * time.Second)
	var final *CertFixJob
	for time.Now().Before(deadline) {
		final, err = GetJob(job.ID)
		if err != nil {
			t.Fatalf("GetJob error = %v", err)
		}
		if final.Status == "completed" || final.Status == "completed_with_failures" {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	if final.Status != "completed_with_failures" {
		t.Fatalf("expected completed_with_failures, got %q", final.Status)
	}
	if final.Failed != 1 {
		t.Fatalf("expected 1 failure, got %d", final.Failed)
	}
}
