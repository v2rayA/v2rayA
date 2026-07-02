package where

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/v2rayA/v2rayA/conf"
)

// Variant identifies the core binary type.
// Since v2rayA now only supports v2raya_core, this is always V2rayaCore.
type Variant string

const (
	// V2rayaCore is the merged v2raya-core binary (xray-core + MultiObservatory).
	// Binary name: v2raya_core
	V2rayaCore Variant = "V2rayaCore"
)

var NotFoundErr = fmt.Errorf("not found")
var ServiceNameList = []string{"v2raya_core"}
var v2rayVersion struct {
	version    string
	binPath    string
	lastUpdate time.Time
	mu         sync.Mutex
}

// GetV2rayServiceVersion returns the version string of the v2raya_core binary.
func GetV2rayServiceVersion() (variant Variant, ver string, err error) {
	// cache for 10 seconds
	v2rayVersion.mu.Lock()
	defer v2rayVersion.mu.Unlock()
	if time.Since(v2rayVersion.lastUpdate) < 10*time.Second {
		return V2rayaCore, v2rayVersion.version, nil
	}

	v2rayPath, err := GetV2rayBinPath()
	if err != nil || len(v2rayPath) <= 0 {
		return V2rayaCore, "", fmt.Errorf("cannot find v2ray executable binary")
	}

	// Get version from binary
	cmd := exec.Command(v2rayPath, "version")
	output := bytes.NewBuffer(nil)
	cmd.Stdout = output
	cmd.Stderr = output
	go func() {
		time.Sleep(5 * time.Second)
		p := cmd.Process
		if p != nil {
			_ = p.Kill()
		}
	}()
	if err := cmd.Start(); err != nil {
		return V2rayaCore, "", err
	}
	cmd.Wait()

	fields := strings.Fields(strings.TrimSpace(output.String()))
	if len(fields) < 2 {
		return V2rayaCore, "", fmt.Errorf("cannot parse version from output: %q", output.String())
	}
	ver = fields[1]

	v2rayVersion.version = ver
	v2rayVersion.binPath = v2rayPath
	v2rayVersion.lastUpdate = time.Now()
	return V2rayaCore, ver, nil
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
	// 先从可执行文件同目录下查找
	if exe, err := os.Executable(); err == nil {
		pa := filepath.Join(filepath.Dir(exe), target)
		if _, err := os.Stat(pa); err == nil {
			return pa, nil
		}
	}
	// 再从 PATH 环境变量里查找
	if pa, err := exec.LookPath(target); err == nil {
		return pa, nil
	}
	return "", NotFoundErr
}
