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

type Variant string

const (
	Unknown Variant = "Unknown"
	V2ray   Variant = "V2Ray"
	Xray    Variant = "Xray"
)

var NotFoundErr = fmt.Errorf("not found")
var ServiceNameList = []string{"xray", "v2ray"}
var v2rayVersion struct {
	variant    Variant
	version    string
	binPath    string
	lastUpdate time.Time
	mu         sync.Mutex
}

/* Detect core type by binary name */
func DetectCoreTypeByBinaryName(binPath string) Variant {
	baseName := strings.ToLower(filepath.Base(binPath))
	// Remove .exe suffix on Windows
	baseName = strings.TrimSuffix(baseName, ".exe")

	switch baseName {
	case "v2ray":
		return V2ray
	case "xray":
		return Xray
	default:
		return Unknown
	}
}

/* get the version of v2ray-core without 'v' like 4.23.1 */
func GetV2rayServiceVersion() (variant Variant, ver string, err error) {
	// cache for 10 seconds
	v2rayVersion.mu.Lock()
	defer v2rayVersion.mu.Unlock()
	if time.Since(v2rayVersion.lastUpdate) < 10*time.Second {
		return v2rayVersion.variant, v2rayVersion.version, nil
	}

	envConfig := conf.GetEnvironmentConfig()
	v2rayPath, err := GetV2rayBinPath()
	if err != nil || len(v2rayPath) <= 0 {
		return Unknown, "", fmt.Errorf("cannot find v2ray executable binary")
	}

	// If user manually specified the binary path, they must also specify the core type
	if envConfig.V2rayBin != "" && envConfig.CoreType == "" {
		return Unknown, "", fmt.Errorf("when using custom v2ray-bin path, you must specify --core-type (v2ray or xray) or set V2RAYA_CORE_TYPE environment variable")
	}

	// Use user-specified core type if provided
	if envConfig.CoreType != "" {
		coreType := strings.ToLower(envConfig.CoreType)
		switch coreType {
		case "v2ray":
			variant = V2ray
		case "xray":
			variant = Xray
		default:
			return Unknown, "", fmt.Errorf("invalid core type '%s', must be 'v2ray' or 'xray'", envConfig.CoreType)
		}
	} else {
		// Auto-detect by binary name
		variant = DetectCoreTypeByBinaryName(v2rayPath)
		if variant == Unknown {
			return Unknown, "", fmt.Errorf("cannot determine core type from binary name '%s', please specify --core-type parameter", v2rayPath)
		}
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
		return Unknown, "", err
	}
	cmd.Wait()

	var fields []string
	if fields = strings.Fields(strings.TrimSpace(output.String())); len(fields) < 2 {
		return Unknown, "", fmt.Errorf("cannot parse version of v2ray")
	}
	ver = fields[1]

	// Verify the detected/specified variant matches the actual binary
	detectedVariant := Unknown
	switch strings.ToUpper(fields[0]) {
	case "V2RAY":
		detectedVariant = V2ray
	case "XRAY":
		detectedVariant = Xray
	}

	if detectedVariant != Unknown && detectedVariant != variant {
		return Unknown, "", fmt.Errorf("core type mismatch: specified/detected '%s' but binary reports '%s'", variant, detectedVariant)
	}

	v2rayVersion.variant = variant
	v2rayVersion.version = ver
	v2rayVersion.binPath = v2rayPath
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
	pa = filepath.Join(pwd, target)
	if _, err := os.Stat(pa); err == nil {
		return pa, nil
	}
	return "", NotFoundErr
}
