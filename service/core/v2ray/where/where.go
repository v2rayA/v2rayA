package where

import (
	"fmt"
	"github.com/v2rayA/v2rayA/conf"
	"os"
	"os/exec"
	"path"
	"runtime"
	"strings"
	"sync"
	"time"
)

var NotFoundErr = fmt.Errorf("not found")
var ServiceNameList = []string{"v2ray"}
var v2rayVersion struct {
	version    string
	lastUpdate time.Time
	mu         sync.Mutex
}

type onceWriter struct {
	buf      []byte
	callback func(buf []byte)
}

func newOnceWriter(callback func(buf []byte)) *onceWriter {
	return &onceWriter{
		callback: callback,
	}
}

func (r *onceWriter) Write(p []byte) (n int, err error) {
	defer func() {
		go r.callback(r.buf)
	}()
	r.buf = make([]byte, len(p))
	copy(r.buf, p)
	return len(p), nil
}

/* get the version of v2ray-core without 'v' like 4.23.1 */
func GetV2rayServiceVersion() (ver string, err error) {
	// cache for 10 seconds
	v2rayVersion.mu.Lock()
	defer v2rayVersion.mu.Unlock()
	if time.Since(v2rayVersion.lastUpdate) < 10*time.Second {
		return v2rayVersion.version, nil
	}
	v2rayPath, err := GetV2rayBinPath()
	if err != nil || len(v2rayPath) <= 0 {
		return "", fmt.Errorf("cannot find v2ray executable binary")
	}
	var output []byte
	var done = make(chan struct{}, 2)
	cmd := exec.Command(v2rayPath, "version")
	cmd.Stdout = newOnceWriter(func(buf []byte) {
		output = buf
		done <- struct{}{}
	})
	cmd.Stderr = cmd.Stdout
	if err := cmd.Start(); err != nil {
		return "", err
	}
	go func() {
		time.Sleep(3 * time.Second)
		_ = cmd.Process.Kill()
		done <- struct{}{}
	}()
	<-done
	var fields []string
	if fields = strings.Fields(strings.TrimSpace(string(output))); len(fields) < 2 {
		return "", fmt.Errorf("cannot parse version of v2ray")
	}
	ver = fields[1]
	if strings.ToUpper(fields[0]) != "V2RAY" {
		ver = "UnknownClient"
	}
	v2rayVersion.version = ver
	v2rayVersion.lastUpdate = time.Now()
	return
}

func GetV2rayBinPath() (string, error) {
	v2rayBinPath := conf.GetEnvironmentConfig().V2rayBin
	if v2rayBinPath == "" {
		return getV2rayBinPathAnyway()
	}
	return v2rayBinPath, nil
}

func getV2rayBinPathAnyway() (path string, err error) {
	for _, target := range ServiceNameList {
		if path, err = getV2rayBinPath(target); err == nil {
			return
		}
	}
	return
}

func getV2rayBinPath(target string) (string, error) {
	if runtime.GOOS == "windows" && !strings.HasSuffix(strings.ToLower(target), ".exe") {
		target += ".exe"
	}
	var pa string
	//从环境变量里找
	pa, err := exec.LookPath(target)
	if err == nil {
		return pa, nil
	}
	//从 pwd 里找
	pwd, err := os.Getwd()
	if err != nil {
		return "", NotFoundErr
	}
	pa = path.Join(pwd, target)
	if _, err := os.Stat(pa); err == nil {
		return pa, nil
	}
	return "", NotFoundErr
}
