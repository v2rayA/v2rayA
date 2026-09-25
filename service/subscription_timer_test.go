package main

import (
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/server/service"
)

func TestSubscriptionTimerSelectsHealthyServer(t *testing.T) {
	if os.Getenv("V2RAYA_V2RAY_BIN") == "" {
		t.Skip("set V2RAYA_V2RAY_BIN to run real-core timer test")
	}
	args := os.Args
	os.Args = os.Args[:1]
	p := *conf.GetEnvironmentConfig()
	os.Args = args
	p.Config = t.TempDir()
	p.Lite = true
	p.CoreStartupTimeout = 2
	conf.SetConfig(p)
	defer db.DB().Close()
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, rw, err := w.(http.Hijacker).Hijack()
		if err != nil {
			return
		}
		defer conn.Close()
		fmt.Fprint(rw, "HTTP/1.1 200 Connection established\r\n\r\n")
		rw.Flush()
		if _, err = http.ReadRequest(rw.Reader); err != nil {
			return
		}
		fmt.Fprint(rw, "HTTP/1.1 204 No Content\r\nContent-Length: 0\r\nConnection: close\r\n\r\n")
		rw.Flush()
	}))
	defer proxy.Close()
	closed, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	badPort := closed.Addr().(*net.TCPAddr).Port
	closed.Close()
	body := fmt.Sprintf("http-proxy://127.0.0.1:%d#dead\nhttp-proxy://127.0.0.1:%d#healthy", badPort, proxy.Listener.Addr().(*net.TCPAddr).Port)
	var timerStarted atomic.Bool
	subscription := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if timerStarted.Swap(false) {
			conf.TickerUpdateSubscription.Reset(time.Hour)
		}
		fmt.Fprint(w, base64.StdEncoding.EncodeToString([]byte(body)))
	}))
	defer subscription.Close()
	if err := configure.SetConfigure(configure.New()); err != nil {
		t.Fatal(err)
	}
	if err := service.Import(subscription.URL, nil); err != nil {
		t.Fatal(err)
	}
	sub := configure.GetSubscription(0)
	sub.AutoSelect = true
	if err := configure.SetSubscription(0, sub); err != nil {
		t.Fatal(err)
	}
	if err := configure.SetOutboundSetting("proxy", configure.OutboundSetting{ProbeURL: "http://probe.invalid/", ProbeInterval: "10s", Type: configure.LeastPing}); err != nil {
		t.Fatal(err)
	}
	initUpdatingTicker()
	defer conf.TickerUpdateGFWList.Stop()
	defer conf.TickerUpdateSubscription.Stop()
	timerStarted.Store(true)
	conf.TickerUpdateSubscription.Reset(50 * time.Millisecond)
	deadline := time.After(5 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("timer did not select healthy node")
		case <-time.After(20 * time.Millisecond):
			service.ConfigurationMu.Lock()
			ws := configure.GetConnectedServersByOutbound("proxy").Get()
			if len(ws) == 1 && ws[0].ID == 2 {
				conf.TickerUpdateSubscription.Stop()
				service.ConfigurationMu.Unlock()
				return
			}
			service.ConfigurationMu.Unlock()
		}
	}
}
