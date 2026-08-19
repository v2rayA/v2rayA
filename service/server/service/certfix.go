package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/matoous/go-nanoid"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/certpin"
	"github.com/v2rayA/v2rayA/kernel/v2ray"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

const (
	CertFixProduct      = "certfix"
	CertFixMaxParallel  = 4
	CertFixProbeTimeout = 10 * time.Second
)

// CertFixCandidate identifies a node that may need its TLS certificate
// configuration adjusted.
type CertFixCandidate struct {
	Which  configure.Which `json:"which"`
	Name   string          `json:"name"`
	Reason string          `json:"reason"`
}

// CertFixNodeResult reports the outcome for a single node.
type CertFixNodeResult struct {
	Which   configure.Which `json:"which"`
	Name    string          `json:"name"`
	Status  string          `json:"status"`
	PinHash string          `json:"pinHash,omitempty"`
	Error   string          `json:"error,omitempty"`
}

// CertFixJob holds the state of a running or completed cert-fix batch.
type CertFixJob struct {
	ID        string              `json:"id"`
	Status    string              `json:"status"`
	Total     int                 `json:"total"`
	Processed int                 `json:"processed"`
	Succeeded int                 `json:"succeeded"`
	Failed    int                 `json:"failed"`
	Results   []CertFixNodeResult `json:"results"`
	Logs      []string            `json:"logs"`
	CreatedAt time.Time           `json:"createdAt"`
	UpdatedAt time.Time           `json:"updatedAt"`

	mu       sync.RWMutex
	saveMu   sync.Mutex
	cancel   context.CancelFunc
	resolved map[configure.Which]*configure.ServerRaw
}

var (
	certFixRegistry = make(map[string]*CertFixJob)
	certFixMu       sync.Mutex
	activeCertFix   *CertFixJob

	// certFixProbe is the probing function used by the fix job. It is
	// overridable in tests so that real network I/O can be avoided.
	certFixProbe = certpin.Probe
)

func init() {
	v2ray.ApiFeed.RegisterProduct(CertFixProduct)
}

// DetectCandidates scans the provided node identifiers and returns those that
// are TLS-bearing and do not already carry an explicit certificate pin.
func DetectCandidates(whiches []*configure.Which) ([]CertFixCandidate, error) {
	var candidates []CertFixCandidate
	for _, w := range whiches {
		if w == nil {
			continue
		}
		sr, err := w.LocateServerRaw()
		if err != nil {
			log.Warn("DetectCandidates: locate failed for %v: %v", w, err)
			continue
		}
		if sr.ServerObj == nil {
			continue
		}
		if !certpin.IsAtRisk(sr.ServerObj) {
			continue
		}
		candidates = append(candidates, CertFixCandidate{
			Which:  *w,
			Name:   sr.ServerObj.GetName(),
			Reason: "tls_without_pin",
		})
	}
	return candidates, nil
}

// StartFix begins a cert-fix job for the supplied candidates. It probes each
// node and applies the appropriate pin or clears insecure fields. Only one job
// may run at a time.
func StartFix(candidates []CertFixCandidate) (*CertFixJob, error) {
	certFixMu.Lock()
	defer certFixMu.Unlock()

	if activeCertFix != nil && activeCertFix.Status == "running" {
		return nil, fmt.Errorf("a cert-fix job is already running")
	}

	id, err := gonanoid.Nanoid(8)
	if err != nil {
		return nil, fmt.Errorf("generate job id: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	job := &CertFixJob{
		ID:        id,
		Status:    "running",
		Total:     len(candidates),
		Results:   make([]CertFixNodeResult, 0, len(candidates)),
		Logs:      make([]string, 0, len(candidates)*2),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		cancel:    cancel,
	}
	certFixRegistry[id] = job
	activeCertFix = job

	go job.run(ctx, candidates)

	return job, nil
}

// GetJob returns the current state of a cert-fix job.
func GetJob(id string) (*CertFixJob, error) {
	certFixMu.Lock()
	job, ok := certFixRegistry[id]
	certFixMu.Unlock()
	if !ok {
		return nil, fmt.Errorf("cert-fix job not found")
	}
	job.mu.RLock()
	defer job.mu.RUnlock()
	return job.snapshot(), nil
}

// CancelJob cancels a running cert-fix job.
func CancelJob(id string) error {
	certFixMu.Lock()
	job, ok := certFixRegistry[id]
	certFixMu.Unlock()
	if !ok {
		return fmt.Errorf("cert-fix job not found")
	}
	job.mu.Lock()
	if job.Status != "running" {
		job.mu.Unlock()
		return fmt.Errorf("job is not running")
	}
	job.Status = "cancelled"
	job.cancel()
	job.mu.Unlock()
	return nil
}

func (job *CertFixJob) run(ctx context.Context, candidates []CertFixCandidate) {
	defer func() {
		certFixMu.Lock()
		if activeCertFix == job {
			activeCertFix = nil
		}
		certFixMu.Unlock()
	}()

	job.resolved = make(map[configure.Which]*configure.ServerRaw, len(candidates))
	for i := range candidates {
		sr, err := candidates[i].Which.LocateServerRaw()
		if err != nil {
			log.Warn("CertFix: locate failed for %v: %v", candidates[i].Which, err)
			continue
		}
		if sr == nil || sr.ServerObj == nil {
			continue
		}
		job.resolved[candidates[i].Which] = sr
	}

	job.logf("Starting cert-fix job for %d node(s)", len(candidates))
	v2ray.ApiFeed.ProductMessage(CertFixProduct, job.feedMessage())

	semaphore := make(chan struct{}, CertFixMaxParallel)
	var wg sync.WaitGroup

	for i := range candidates {
		if ctx.Err() != nil {
			break
		}
		wg.Add(1)
		semaphore <- struct{}{}
		go func(index int) {
			defer wg.Done()
			defer func() { <-semaphore }()
			job.processOne(ctx, &candidates[index])
		}(i)
	}

	wg.Wait()

	var succeeded, failed int
	var terminalStatus string
	job.mu.Lock()
	if job.Status == "running" {
		if job.Failed > 0 {
			job.Status = "completed_with_failures"
		} else {
			job.Status = "completed"
		}
	}
	terminalStatus = job.Status
	job.UpdatedAt = time.Now()
	succeeded = job.Succeeded
	failed = job.Failed
	job.mu.Unlock()

	job.logf("Cert-fix job finished (%s): %d succeeded, %d failed", terminalStatus, succeeded, failed)
	v2ray.ApiFeed.ProductMessage(CertFixProduct, job.feedMessage())
}

func (job *CertFixJob) processOne(ctx context.Context, candidate *CertFixCandidate) {
	result := CertFixNodeResult{
		Which: candidate.Which,
		Name:  candidate.Name,
	}

	sr, ok := job.resolved[candidate.Which]
	if !ok {
		result.Status = "failed"
		result.Error = "server not resolved before job start"
		job.recordResult(result)
		return
	}
	if sr.ServerObj == nil {
		result.Status = "failed"
		result.Error = "server object is nil"
		job.recordResult(result)
		return
	}

	target, ok := certpin.ExtractTarget(sr.ServerObj)
	if !ok {
		result.Status = "skipped"
		result.Error = "not a TLS-bearing outbound"
		job.recordResult(result)
		return
	}

	probeCtx, cancel := context.WithTimeout(ctx, CertFixProbeTimeout)
	res, err := certFixProbe(probeCtx, target)
	cancel()
	if err != nil {
		result.Status = "failed"
		result.Error = "probe: " + err.Error()
		job.recordResult(result)
		return
	}
	if res.Error != "" {
		result.Status = "failed"
		result.Error = res.Error
		job.recordResult(result)
		return
	}

	if res.Trusted {
		result.Status = "trusted"
		job.logf("[%s] certificate is system-trusted", candidate.Name)
	} else {
		result.Status = "pinned"
		if target.HashType == certpin.HashTypeChain {
			result.PinHash = res.ChainHash
		} else {
			result.PinHash = res.LeafHash
		}
		job.logf("[%s] pinned certificate %s...%s", candidate.Name, result.PinHash[:8], result.PinHash[len(result.PinHash)-8:])
	}

	if err := certpin.ApplyResult(sr.ServerObj, res); err != nil {
		result.Status = "failed"
		result.Error = "apply result: " + err.Error()
		job.recordResult(result)
		return
	}

	job.saveMu.Lock()
	err = saveServerRaw(candidate.Which, sr)
	job.saveMu.Unlock()
	if err != nil {
		result.Status = "failed"
		result.Error = "save server: " + err.Error()
		job.recordResult(result)
		return
	}

	job.recordResult(result)
}

func (job *CertFixJob) recordResult(result CertFixNodeResult) {
	job.mu.Lock()
	defer job.mu.Unlock()
	job.Results = append(job.Results, result)
	job.Processed++
	if result.Status == "trusted" || result.Status == "pinned" {
		job.Succeeded++
	} else if result.Status == "failed" {
		job.Failed++
	}
	job.UpdatedAt = time.Now()
	v2ray.ApiFeed.ProductMessage(CertFixProduct, job.feedMessageLocked())
}

func (job *CertFixJob) logf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	job.mu.Lock()
	job.Logs = append(job.Logs, fmt.Sprintf("[%s] %s", time.Now().Format("15:04:05"), msg))
	job.mu.Unlock()
	log.Info("certfix: %s", msg)
}

func (job *CertFixJob) feedMessage() map[string]interface{} {
	job.mu.RLock()
	defer job.mu.RUnlock()
	return job.feedMessageLocked()
}

func (job *CertFixJob) feedMessageLocked() map[string]interface{} {
	return map[string]interface{}{
		"jobId":     job.ID,
		"status":    job.Status,
		"total":     job.Total,
		"processed": job.Processed,
		"succeeded": job.Succeeded,
		"failed":    job.Failed,
		"logs":      job.Logs,
	}
}

func (job *CertFixJob) snapshot() *CertFixJob {
	return &CertFixJob{
		ID:        job.ID,
		Status:    job.Status,
		Total:     job.Total,
		Processed: job.Processed,
		Succeeded: job.Succeeded,
		Failed:    job.Failed,
		Results:   append([]CertFixNodeResult(nil), job.Results...),
		Logs:      append([]string(nil), job.Logs...),
		CreatedAt: job.CreatedAt,
		UpdatedAt: job.UpdatedAt,
	}
}

func saveServerRaw(w configure.Which, sr *configure.ServerRaw) error {
	switch w.TYPE {
	case configure.ServerType:
		return configure.SetServer(w.ID-1, sr)
	case configure.SubscriptionServerType:
		sub := configure.GetSubscription(w.Sub)
		if sub == nil {
			return fmt.Errorf("subscription %d not found", w.Sub)
		}
		sub.Servers[w.ID-1] = *sr
		return configure.SetSubscription(w.Sub, sub)
	default:
		return fmt.Errorf("unsupported which type: %v", w.TYPE)
	}
}
