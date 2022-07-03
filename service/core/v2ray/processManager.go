package v2ray

import (
	"fmt"
	"github.com/v2rayA/v2rayA/common/resolv"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/core/specialMode"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"os/exec"
	"sync"
)

type CoreProcessManager struct {
	p  *Process
	mu sync.Mutex
}

var ProcessManager CoreProcessManager

func (m *CoreProcessManager) beforeStop(p *Process) {
	if p.template.Setting.Transparent != configure.TransparentClose &&
		conf.GetEnvironmentConfig().TransparentHook != "" {
		log.Info("Execute the transparent pre stop hook: %v", conf.GetEnvironmentConfig().TransparentHook)
		b, err := exec.Command(conf.GetEnvironmentConfig().TransparentHook, fmt.Sprintf("--transparent-type=%v", p.template.Setting.TransparentType), "--stage=pre-stop").CombinedOutput()
		if err != nil {
			log.Warn("Error when executing the transparent pre stop hook: %v", err)
			return
		}
		if len(b) > 0 {
			log.Info("Executing the transparent pre stop hook: %v", string(b))
		}
	}
	CheckAndStopTransparentProxy()
	specialMode.StopDNSSupervisor()
}

func (m *CoreProcessManager) afterStop(p *Process) {
	if p.template.Setting.Transparent != configure.TransparentClose &&
		conf.GetEnvironmentConfig().TransparentHook != "" {
		log.Info("Execute the transparent after stop hook: %v", conf.GetEnvironmentConfig().TransparentHook)
		b, err := exec.Command(conf.GetEnvironmentConfig().TransparentHook, fmt.Sprintf("--transparent-type=%v", p.template.Setting.TransparentType), "--stage=post-stop").CombinedOutput()
		if err != nil {
			log.Warn("Error when executing the transparent after stop hook: %v", err)
			return
		}
		if len(b) > 0 {
			log.Info("Executing the transparent after stop hook: %v", string(b))
		}
	}
}

func (m *CoreProcessManager) Stop(saveRunning bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stop(saveRunning)
}

func (m *CoreProcessManager) stop(saveRunning bool) {
	if m.p == nil {
		return
	}

	m.beforeStop(m.p)
	err := m.p.Close()
	if err != nil {
		log.Warn("CoreProcessManager.Stop: %v", err)
	}
	if saveRunning {
		configure.SetRunning(false)
	}

	m.afterStop(m.p)

	m.p = nil
}

func (m *CoreProcessManager) beforeStart(t *Template) (err error) {
	if t.Setting.Transparent != configure.TransparentClose &&
		conf.GetEnvironmentConfig().TransparentHook != "" {
		log.Info("Execute the transparent pre start hook: %v", conf.GetEnvironmentConfig().TransparentHook)
		b, err := exec.Command(conf.GetEnvironmentConfig().TransparentHook, fmt.Sprintf("--transparent-type=%v", t.Setting.TransparentType), "--stage=pre-start").CombinedOutput()
		if err != nil {
			return fmt.Errorf("executing the transparent pre start hook: %w", err)
		}
		if len(b) > 0 {
			log.Info("Executing the transparent pre start hook: %v", string(b))
		}
	}

	resolv.CheckResolvConf()

	if (t.Setting.Transparent == configure.TransparentGfwlist || t.Setting.RulePortMode == configure.GfwlistMode) && !asset.DoesV2rayAssetExist("LoyalsoldierSite.dat") {
		return fmt.Errorf("cannot find GFWList files. update GFWList and try again")
	}

	if err = t.CheckInboundPortsOccupied(); err != nil {
		return err
	}

	return nil
}

func (m *CoreProcessManager) afterStart(t *Template) (err error) {
	if err = CheckAndSetupTransparentProxy(true, t.Setting); err != nil {
		m.stop(true)
		return
	}
	specialMode.CheckAndSetupDNSSupervisor()

	if t.Setting.Transparent != configure.TransparentClose &&
		conf.GetEnvironmentConfig().TransparentHook != "" {
		log.Info("Execute the transparent after start hook: %v", conf.GetEnvironmentConfig().TransparentHook)
		b, err := exec.Command(conf.GetEnvironmentConfig().TransparentHook, fmt.Sprintf("--transparent-type=%v", t.Setting.TransparentType), "--stage=post-start").CombinedOutput()
		if err != nil {
			return fmt.Errorf("executing the transparent after start hook: %w", err)
		}
		if len(b) > 0 {
			log.Info("Executing the transparent after start hook: %v", string(b))
		}
	}
	return nil
}

func (m *CoreProcessManager) Start(t *Template) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stop(true)

	if err := m.beforeStart(t); err != nil {
		return err
	}

	process, err := NewProcess(t)
	if err != nil {
		return err
	}
	m.p = process

	if err := m.afterStart(t); err != nil {
		return err
	}

	configure.SetRunning(true)
	return nil
}

// Running reports if v2ray-core is running.
func (m *CoreProcessManager) Running() bool {
	return m.p != nil
}

func (m *CoreProcessManager) Process() *Process {
	return m.p
}
