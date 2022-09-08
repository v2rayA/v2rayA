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
	m.CheckAndStopTransparentProxy(p.template.Setting)
	specialMode.StopDNSSupervisor()

	if conf.GetEnvironmentConfig().CoreHook != "" {
		log.Info("Execute the core pre stop hook: %v", conf.GetEnvironmentConfig().CoreHook)
		b, err := exec.Command(conf.GetEnvironmentConfig().CoreHook,
			"--stage=pre-stop",
			fmt.Sprintf("--v2raya-confdir=%v", conf.GetEnvironmentConfig().Config),
		).CombinedOutput()
		if len(b) > 0 {
			log.Info("Executing the core pre stop hook: %v", string(b))
		}
		if err != nil {
			log.Warn("Error when executing the core pre stop hook: %v", err)
			return
		}
	}
}

func (m *CoreProcessManager) GetRunningTemplate() *Template {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.p == nil {
		return nil
	}
	return m.p.template
}

func (m *CoreProcessManager) CheckAndSetupTransparentProxy(checkRunning bool, setting *configure.Setting) (err error) {
	if setting != nil {
		setting.FillEmpty()
	} else {
		setting = configure.GetSettingNotNil()
	}
	if (!checkRunning || ProcessManager.Running()) && IsTransparentOn() {
		deleteTransparentProxyRules()

		if conf.GetEnvironmentConfig().TransparentHook != "" {
			log.Info("Execute the transparent pre start hook: %v", conf.GetEnvironmentConfig().TransparentHook)
			b, err := exec.Command(conf.GetEnvironmentConfig().TransparentHook,
				fmt.Sprintf("--transparent-type=%v", setting.TransparentType),
				"--stage=pre-start",
				fmt.Sprintf("--v2raya-confdir=%v", conf.GetEnvironmentConfig().Config),
			).CombinedOutput()
			if len(b) > 0 {
				log.Info("Executing the transparent pre start hook: %v", string(b))
			}
			if err != nil {
				return fmt.Errorf("error when executing the transparent pre start hook: %w", err)
			}
		}

		err = writeTransparentProxyRules()

		if conf.GetEnvironmentConfig().TransparentHook != "" {
			log.Info("Execute the transparent post start hook: %v", conf.GetEnvironmentConfig().TransparentHook)
			b, err := exec.Command(conf.GetEnvironmentConfig().TransparentHook,
				fmt.Sprintf("--transparent-type=%v", setting.TransparentType),
				"--stage=post-start",
				fmt.Sprintf("--v2raya-confdir=%v", conf.GetEnvironmentConfig().Config),
			).CombinedOutput()
			if len(b) > 0 {
				log.Info("Executing the transparent post start hook: %v", string(b))
			}
			if err != nil {
				return fmt.Errorf("error when executing the transparent post start hook: %w", err)
			}
		}
	}
	return
}

func (m *CoreProcessManager) CheckAndStopTransparentProxy(setting *configure.Setting) {
	if setting == nil {
		if t := m.GetRunningTemplate(); t != nil {
			setting = t.Setting
		} else {
			return
		}
	}
	if setting.Transparent != configure.TransparentClose {
		if conf.GetEnvironmentConfig().TransparentHook != "" {
			log.Info("Execute the transparent pre stop hook: %v", conf.GetEnvironmentConfig().TransparentHook)
			b, err := exec.Command(conf.GetEnvironmentConfig().TransparentHook,
				fmt.Sprintf("--transparent-type=%v", setting.TransparentType),
				"--stage=pre-stop",
				fmt.Sprintf("--v2raya-confdir=%v", conf.GetEnvironmentConfig().Config),
			).CombinedOutput()
			if len(b) > 0 {
				log.Info("Executing the transparent pre stop hook: %v", string(b))
			}
			if err != nil {
				log.Warn("Error when executing the transparent pre stop hook: %v", err)
				return
			}
		}

		deleteTransparentProxyRules()

		if conf.GetEnvironmentConfig().TransparentHook != "" {
			log.Info("Execute the transparent post stop hook: %v", conf.GetEnvironmentConfig().TransparentHook)
			b, err := exec.Command(conf.GetEnvironmentConfig().TransparentHook,
				fmt.Sprintf("--transparent-type=%v", setting.TransparentType),
				"--stage=post-stop",
				fmt.Sprintf("--v2raya-confdir=%v", conf.GetEnvironmentConfig().Config),
			).CombinedOutput()
			if len(b) > 0 {
				log.Info("Executing the transparent post stop hook: %v", string(b))
			}
			if err != nil {
				log.Warn("Error when executing the transparent post stop hook: %v", err)
				return
			}
		}
	}
}

func (m *CoreProcessManager) afterStop(p *Process) {
	if conf.GetEnvironmentConfig().CoreHook != "" {
		log.Info("Execute the core post stop hook: %v", conf.GetEnvironmentConfig().CoreHook)
		b, err := exec.Command(conf.GetEnvironmentConfig().CoreHook,
			"--stage=post-stop",
			fmt.Sprintf("--v2raya-confdir=%v", conf.GetEnvironmentConfig().Config),
		).CombinedOutput()
		if len(b) > 0 {
			log.Info("Executing the core post stop hook: %v", string(b))
		}
		if err != nil {
			log.Warn("Error when executing the core post stop hook: %v", err)
			return
		}
	}
}

func (m *CoreProcessManager) Stop(saveRunning bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stop(saveRunning)
	if m.p != nil {
		for _, pm := range m.p.pluginManagers {
			_ = pm.Kill()
		}
		m.p.pluginManagers = nil
	}
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
	resolv.CheckResolvConf()

	if (t.Setting.Transparent == configure.TransparentGfwlist || t.Setting.RulePortMode == configure.GfwlistMode) && !asset.DoesV2rayAssetExist("LoyalsoldierSite.dat") {
		return fmt.Errorf("cannot find GFWList files. update GFWList and try again")
	}

	if err = t.CheckInboundPortsOccupied(); err != nil {
		return err
	}

	if conf.GetEnvironmentConfig().CoreHook != "" {
		log.Info("Execute the core pre start hook: %v", conf.GetEnvironmentConfig().CoreHook)
		b, err := exec.Command(conf.GetEnvironmentConfig().CoreHook,
			"--stage=pre-start",
			fmt.Sprintf("--v2raya-confdir=%v", conf.GetEnvironmentConfig().Config),
		).CombinedOutput()
		if len(b) > 0 {
			log.Info("Executing the core pre start hook: %v", string(b))
		}
		if err != nil {
			return fmt.Errorf("error when executing the core pre start hook: %w", err)
		}
	}
	return nil
}

func (m *CoreProcessManager) afterStart(t *Template) (err error) {
	if err = m.CheckAndSetupTransparentProxy(false, t.Setting); err != nil {
		return err
	}
	specialMode.CheckAndSetupDNSSupervisor()

	if conf.GetEnvironmentConfig().CoreHook != "" {
		log.Info("Execute the core post start hook: %v", conf.GetEnvironmentConfig().CoreHook)
		b, err := exec.Command(conf.GetEnvironmentConfig().CoreHook,
			"--stage=post-start",
			fmt.Sprintf("--v2raya-confdir=%v", conf.GetEnvironmentConfig().Config),
		).CombinedOutput()
		if len(b) > 0 {
			log.Info("Executing the core post start hook: %v", string(b))
		}
		if err != nil {
			return fmt.Errorf("error when executing the core post start hook: %w", err)
		}
	}
	return nil
}

func (m *CoreProcessManager) Start(t *Template) (err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stop(true)

	process, err := NewProcess(t, func() error {
		return m.beforeStart(t)
	}, func() error {
		return m.afterStart(t)
	})
	if err != nil {
		return err
	}
	m.p = process
	defer func() {
		if err != nil {
			m.stop(true)
		}
	}()

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
