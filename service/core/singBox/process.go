package singBox

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/core/v2ray/where"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

var ServiceName = "sing-box"

// Process is a sing-box process
type Process struct {
	// mutex protect the proc
	mutex      sync.Mutex
	proc       *os.Process
	procCancel func() // cancel func for proc
	template   *Template
}

var singBoxVersion struct {
	version    string
	lastUpdate time.Time
	mu         sync.Mutex
}

func GetSingBoxVersion() (ver string, err error) {
	// cache for 10 seconds
	singBoxVersion.mu.Lock()
	defer singBoxVersion.mu.Unlock()
	if time.Since(singBoxVersion.lastUpdate) < 10*time.Second {
		return singBoxVersion.version, nil
	}
	singBoxPath, err := GetSingBoxBinPath()
	if err != nil || len(singBoxPath) <= 0 {
		return "", fmt.Errorf("cannot find sing-box executable binary")
	}
	cmd := exec.Command(singBoxPath, "version")
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
		return "", err
	}
	cmd.Wait()
	var fields []string
	if fields = strings.Fields(strings.TrimSpace(output.String())); len(fields) < 3 {
		return "", fmt.Errorf("cannot parse version of sing-box")
	}
	ver = fields[2]
	singBoxVersion.version = ver
	singBoxVersion.lastUpdate = time.Now()
	return
}

func GetSingBoxBinPath() (string, error) {
	target := ServiceName
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
		return "", where.NotFoundErr
	}
	pa = filepath.Join(pwd, target)
	if _, err := os.Stat(pa); err == nil {
		return pa, nil
	}
	return "", where.NotFoundErr
}

func GetSingBoxConfigPath() (p string) {
	return path.Join(conf.GetEnvironmentConfig().Config, "tun_config.json")
}

func NewProcess(tmpl *Template,
	prestart func() error, poststart func() error,
	postUnexpectedStop func(p *Process),
) (*Process, error) {
	process := &Process{
		template: tmpl,
	}
	err := WriteSingBoxConfig(tmpl.ToConfigBytes())
	if err != nil {
		return nil, err
	}
	pCtx, cancel := context.WithCancel(context.Background())
	defer func() {
		if err != nil {
			cancel()
		}
	}()
	process.procCancel = cancel
	if err = prestart(); err != nil {
		return nil, err
	}
	proc, err := StartCoreProcess(pCtx)
	if err != nil {
		return nil, err
	}
	if err = poststart(); err != nil {
		return nil, err
	}
	process.proc = proc
	go func() {
		p, e := proc.Wait()
		if process.procCancel == nil {
			// canceled by v2rayA
			return
		}
		defer postUnexpectedStop(process)
		var t []string
		if p != nil {
			if p.Success() {
				return
			}
			t = append(t, p.String())
		}
		if e != nil {
			t = append(t, e.Error())
		}
		log.Warn("sing-box: %v", strings.Join(t, ": "))
	}()
	return process, nil
}

type logInfoWriter struct {
}

func (w logInfoWriter) Write(p []byte) (n int, err error) {
	s := string(p)
	// trim the ending \n
	length := len(s)
	if s[length-1] == '\n' {
		s = s[:length-1]
	}
	// print each line separately
	lines := strings.Split(s, "\n")
	for _, line := range lines {
		// remove timestamp
		fields := strings.SplitN(line, " ", 3)
		if len(fields) >= 3 {
			if _, err := time.Parse("2006/01/02 15:04:05", fields[0]+" "+fields[1]); err == nil {
				log.Info("%v", fields[2])
			} else {
				log.Info("%v", line)
			}
		} else {
			log.Info("%v", line)
		}

	}
	return len(p), nil
}

var logWriter logInfoWriter

func (p *Process) Close() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	if p.procCancel != nil {
		p.procCancel()
		p.procCancel = nil
		return nil
	} else {
		_, err := p.proc.Wait()
		return err
	}
}

func RunWithLog(ctx context.Context, name string, argv []string, dir string, env []string) (*os.Process, error) {
	cmd := exec.CommandContext(ctx, name)
	cmd.Args = argv
	cmd.Dir = dir
	cmd.Env = env
	cmd.Stdout = logWriter
	cmd.Stderr = logWriter
	err := cmd.Start()
	if err != nil {
		return nil, err
	}
	return cmd.Process, nil
}

func StartCoreProcess(ctx context.Context) (*os.Process, error) {
	singBoxBinPath, err := GetSingBoxBinPath()
	if err != nil {
		return nil, err
	}
	v2rayPath, err := where.GetV2rayBinPath()
	if err != nil {
		return nil, err
	}
	dir := filepath.Dir(v2rayPath)
	var arguments = []string{
		singBoxBinPath,
		"run",
		"-c",
		GetSingBoxConfigPath(),
	}
	log.Debug(strings.Join(arguments, " "))
	proc, err := RunWithLog(ctx, singBoxBinPath, arguments, dir, nil)
	if err != nil {
		return nil, err
	}
	return proc, nil
}
