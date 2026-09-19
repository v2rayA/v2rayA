package v2ray

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/v2rayA/v2rayA/common/resolv"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/v2ray/asset"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

type CoreProcessManager struct {
	shuttingDown     atomic.Bool
	p                *Process
	startMu          sync.Mutex
	mu               sync.Mutex
	generation       uint64
	testing          bool
	networkPaused    bool
	connectivityStop chan struct{}
	// transparentOn records whether this process installed transparent proxy
	// rules. Shutdown asks for the teardown twice — once in pre_run and once
	// through Stop — which ran the user's pre-stop and post-stop hooks twice.
	transparentOn atomic.Bool
}

var ProcessManager CoreProcessManager

func (m *CoreProcessManager) beforeStop(p *Process) {
	hostMu.Lock()
	m.checkAndStopTransparentProxy(p.template.Setting)
	hostMu.Unlock()

	if corehook := conf.GetEnvironmentConfig().CoreHook; corehook != "" {
		hook := strings.Split(corehook, " ")
		hook = append(hook, "--stage=pre-stop", fmt.Sprintf("--v2raya-confdir=%v", conf.GetEnvironmentConfig().Config))
		log.Info("Execute the core pre stop hook: %v", hook)
		b, err := exec.Command(hook[0], hook[1:]...).CombinedOutput()
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

func (m *CoreProcessManager) SetLatencyTesting(testing bool) {
	m.mu.Lock()
	m.testing = testing
	m.mu.Unlock()
}

// ServiceRunning reports whether the proxy service should be treated as running
// by the frontend. Temporary latency tests and connectivity-paused transparent
// proxy states are excluded from this view.
func (m *CoreProcessManager) ServiceRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.p != nil && !m.testing && !m.networkPaused
}

func (m *CoreProcessManager) NetworkPaused() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.networkPaused
}

func (m *CoreProcessManager) stopConnectivityMonitorLocked() {
	if m.connectivityStop != nil {
		close(m.connectivityStop)
		m.connectivityStop = nil
	}
}

func (m *CoreProcessManager) CheckAndSetupTransparentProxy(checkRunning bool, setting *configure.Setting, tmpl *Template) (err error) {
	m.mu.Lock()
	p, generation := m.p, m.generation
	if p == nil && checkRunning {
		m.mu.Unlock()
		return nil
	}
	if p == nil || p.template != tmpl {
		m.mu.Unlock()
		return fmt.Errorf("core process changed before transparent setup")
	}
	m.mu.Unlock()
	return m.setupTransparentProxy(p, generation, setting, tmpl)
}

func (m *CoreProcessManager) setupTransparentProxy(p *Process, generation uint64, setting *configure.Setting, tmpl *Template) (err error) {
	if setting != nil {
		setting.FillEmpty()
	} else {
		setting = configure.GetSettingNotNil()
	}
	if IsTransparentOn(setting) {
		if err = m.mutateHost(p, generation, func() error { deleteTransparentProxyRules(); return nil }); err != nil {
			return err
		}

		runHook := func(stage string) error {
			thook := conf.GetEnvironmentConfig().TransparentHook
			if thook == "" {
				return nil
			}
			hook := strings.Split(thook, " ")
			hook = append(hook,
				fmt.Sprintf("--transparent-type=%v", setting.TransparentType),
				"--stage="+stage,
				fmt.Sprintf("--v2raya-confdir=%v", conf.GetEnvironmentConfig().Config))
			log.Info("Execute the transparent %s hook: %v", stage, hook)
			b, err := exec.Command(hook[0], hook[1:]...).CombinedOutput()
			if len(b) > 0 {
				log.Info("Executing the transparent %s hook: %v", stage, string(b))
			}
			if err != nil {
				return fmt.Errorf("error when executing the transparent %s hook: %w", stage, err)
			}
			return nil
		}
		if err = m.mutateHost(p, generation, func() error { return runHook("pre-start") }); err != nil {
			return err
		}
		waitForTransparentDNS(tmpl)
		if err = m.checkProcessOwner(p, generation); err != nil {
			return err
		}
		if err = m.mutateHost(p, generation, func() error {
			// Partial rulesets must also be torn down.
			m.transparentOn.Store(true)
			return writeTransparentProxyRules(tmpl)
		}); err != nil {
			return err
		}
		return m.mutateHost(p, generation, func() error { return runHook("post-start") })
	}
	return
}

func (m *CoreProcessManager) CheckAndStopTransparentProxy(setting *configure.Setting) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if setting == nil && m.p != nil {
		setting = m.p.template.Setting
	}
	hostMu.Lock()
	defer hostMu.Unlock()
	m.checkAndStopTransparentProxy(setting)
}

func (m *CoreProcessManager) checkAndStopTransparentProxy(setting *configure.Setting) {
	if !m.transparentOn.Swap(false) {
		// Nothing installed by this process, so nothing to remove and no hook
		// to run. Rules left behind by an earlier process are removed by
		// cleanupResidualTransparentProxyRules when the rules are set up.
		return
	}
	if setting == nil {
		m.transparentOn.Store(true)
		return
	}
	if setting.Transparent != configure.TransparentClose {
		if thook := conf.GetEnvironmentConfig().TransparentHook; thook != "" {
			hook := strings.Split(thook, " ")
			hook = append(hook,
				fmt.Sprintf("--transparent-type=%v", setting.TransparentType),
				"--stage=pre-stop",
				fmt.Sprintf("--v2raya-confdir=%v", conf.GetEnvironmentConfig().Config))
			log.Info("Execute the transparent pre stop hook: %v", hook)
			b, err := exec.Command(hook[0], hook[1:]...).CombinedOutput()
			if len(b) > 0 {
				log.Info("Executing the transparent pre stop hook: %v", string(b))
			}
			if err != nil {
				log.Warn("Error when executing the transparent pre stop hook: %v", err)
			}
		}

		deleteTransparentProxyRules()

		if thook := conf.GetEnvironmentConfig().TransparentHook; thook != "" {
			hook := strings.Split(thook, " ")
			hook = append(hook,
				fmt.Sprintf("--transparent-type=%v", setting.TransparentType),
				"--stage=post-stop",
				fmt.Sprintf("--v2raya-confdir=%v", conf.GetEnvironmentConfig().Config))
			log.Info("Execute the transparent post stop hook: %v", hook)
			b, err := exec.Command(hook[0], hook[1:]...).CombinedOutput()
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
	if corehook := conf.GetEnvironmentConfig().CoreHook; corehook != "" {
		hook := strings.Split(corehook, " ")
		hook = append(hook, "--stage=post-stop", fmt.Sprintf("--v2raya-confdir=%v", conf.GetEnvironmentConfig().Config))
		log.Info("Execute the core post stop hook: %v", hook)
		b, err := exec.Command(hook[0], hook[1:]...).CombinedOutput()
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
	if m.p == nil {
		return
	}
	if saveRunning {
		// User explicitly stopped the core via API (StopV2ray).
		defer func() { _ = configure.SetLastKernelExitStatus(configure.LastKernelExitStopped) }()
	} else {
		// Graceful shutdown — the kernel was running when v2rayA exited.
		// Preserve the "running" status so it is restored on next startup.
		defer func() { _ = configure.SetLastKernelExitStatus(configure.LastKernelExitRunning) }()
	}
	m.stop(saveRunning)
}

func (m *CoreProcessManager) stop(saveRunning bool) {
	if m.p == nil {
		return
	}

	p := m.p

	m.stopConnectivityMonitorLocked()
	m.networkPaused = false

	m.beforeStop(p)

	err := p.Close()
	if err != nil {
		log.Warn("CoreProcessManager.Stop: %v", err)
	}
	if saveRunning {
		configure.SetRunning(false)
	}

	m.afterStop(p)

	m.p = nil

	// Notify connected frontend clients that the proxy is no longer running.
	ApiFeed.ProductMessage("running_state", map[string]interface{}{"running": false, "networkPaused": false})
}

// MarkShuttingDown records that the service is exiting: a core that dies
// now died with it (systemd sends the whole cgroup SIGTERM), not on its own.
func (m *CoreProcessManager) MarkShuttingDown() {
	m.shuttingDown.Store(true)
}

func (m *CoreProcessManager) handleUnexpectedStop(p *Process) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.p != p {
		return
	}
	if m.shuttingDown.Load() {
		// keep running=true and the "running" exit status so the next start
		// restores the core instead of reporting a crash
		m.stop(false)
		return
	}
	m.stop(true)
	// Override the default status (Stop() would have saved "running" or "stopped")
	// to record the abnormal exit so the startup code can warn the user.
	_ = configure.SetLastKernelExitStatus(configure.LastKernelExitCrashed)
}

// runPreStartHook executes the configured core pre-start hook.
// It is called inside the Start lock so it is correctly ordered after the
// pre-stop hook that runs inside m.stop.
func (m *CoreProcessManager) runPreStartHook() error {
	if corehook := conf.GetEnvironmentConfig().CoreHook; corehook != "" {
		hook := strings.Split(corehook, " ")
		hook = append(hook, "--stage=pre-start", fmt.Sprintf("--v2raya-confdir=%v", conf.GetEnvironmentConfig().Config))
		log.Info("Execute the core pre start hook: %v", hook)
		deadline := time.Duration(conf.GetEnvironmentConfig().CoreStartupTimeout) * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), deadline)
		defer cancel()
		cmd := exec.CommandContext(ctx, hook[0], hook[1:]...)
		if setHookProcessGroup(cmd) {
			cmd.Cancel = func() error {
				group, err := os.FindProcess(-cmd.Process.Pid)
				if err != nil {
					return err
				}
				return group.Kill()
			}
		}
		cmd.WaitDelay = 100 * time.Millisecond
		b, err := cmd.CombinedOutput()
		if len(b) > 0 {
			log.Info("Executing the core pre start hook: %v", string(b))
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return fmt.Errorf("core pre-start hook %q exceeded the %s deadline", corehook, deadline)
		}
		if err != nil {
			return fmt.Errorf("error when executing the core pre start hook: %w", err)
		}
	}
	return nil
}

func (m *CoreProcessManager) ownsProcessLocked(p *Process, generation uint64) bool {
	return p != nil && m.p == p && m.generation == generation
}

func (m *CoreProcessManager) checkProcessOwner(p *Process, generation uint64) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.ownsProcessLocked(p, generation) {
		return fmt.Errorf("core process exited or was replaced during startup")
	}
	return nil
}

func (m *CoreProcessManager) mutateHost(p *Process, generation uint64, mutate func() error) error {
	m.mu.Lock()
	if !m.ownsProcessLocked(p, generation) {
		m.mu.Unlock()
		return fmt.Errorf("core process exited or was replaced before host setup")
	}
	hostMu.Lock()
	m.mu.Unlock()
	err := mutate()
	hostMu.Unlock()
	return errors.Join(err, m.checkProcessOwner(p, generation))
}

func (m *CoreProcessManager) afterStart(p *Process, generation uint64) (err error) {
	t := p.template
	if err = m.setupTransparentProxy(p, generation, t.Setting, t); err != nil {
		return err
	}
	m.startConnectivityMonitor(p, generation)
	if err := m.checkProcessOwner(p, generation); err != nil {
		return err
	}

	if corehook := conf.GetEnvironmentConfig().CoreHook; corehook != "" {
		hook := strings.Split(corehook, " ")
		hook = append(hook, "--stage=post-start", fmt.Sprintf("--v2raya-confdir=%v", conf.GetEnvironmentConfig().Config))
		log.Info("Execute the core post start hook: %v", hook)
		b, err := exec.Command(hook[0], hook[1:]...).CombinedOutput()
		if len(b) > 0 {
			log.Info("Executing the core post start hook: %v", string(b))
		}
		if err != nil {
			return fmt.Errorf("error when executing the core post start hook: %w", err)
		}
	}
	return m.checkProcessOwner(p, generation)
}

func (m *CoreProcessManager) Start(t *Template) (err error) {
	m.startMu.Lock()
	defer m.startMu.Unlock()
	// Phase 1 (pre-lock): lightweight checks that do not depend on whether a
	// previous process is running.  Port occupancy is checked by NewProcess
	// after the old process has been stopped.
	resolv.CheckResolvConf()
	if (t.Setting.Transparent == configure.TransparentGfwlist || t.Setting.RulePortMode == configure.GfwlistMode) && !asset.DoesV2rayAssetExist("LoyalsoldierSite.dat") {
		_ = t.Close()
		return asset.GFWListMissingError()
	}

	// Phase 2 (locked): stop the old process, run the pre-start hook (ordered
	// after the pre-stop hook in beforeStop), then start the new core.
	// afterStart is deferred to Phase 3 so that heavy operations (DNS, TUN,
	// transparent-proxy hooks) do not block while the lock is held.
	m.mu.Lock()
	m.stop(true)
	// A marker left by a start that never committed is torn down first. A
	// teardown that fails is logged, not fatal: refusing every later start
	// would leave that state behind with no service to fix it.
	if state, err := configure.GetHostState(); err != nil {
		log.Warn("read pending host state: %v", err)
	} else if state != nil {
		if err := RecoverHostState(state); err != nil {
			log.Warn("recover pending host state: %v", err)
		}
		if err := configure.SetHostState(nil); err != nil {
			log.Warn("clear pending host state: %v", err)
		}
	}
	m.generation++
	generation := m.generation
	process, err := NewProcess(t, func() error {
		return m.runPreStartHook()
	}, func() error {
		return nil // afterStart executed post-lock in Phase 3
	}, m.handleUnexpectedStop)
	if err != nil {
		m.mu.Unlock()
		return err
	}
	m.p = process
	testing := m.testing
	err = configure.SetHostState(&configure.HostState{
		TransparentType:   t.Setting.TransparentType,
		APIPort:           t.ApiPort,
		TunAutoRoute:      t.Setting.TunAutoRoute,
		TunTeardownScript: t.Setting.TunTeardownScript,
	})
	m.mu.Unlock()

	defer func() {
		if err != nil {
			m.mu.Lock()
			defer m.mu.Unlock()
			if m.ownsProcessLocked(process, generation) {
				// stop tears the host state down itself; the marker only
				// records that it happened
				m.stop(true)
				err = errors.Join(err, configure.SetHostState(nil))
			} else if m.generation == generation && m.p == nil {
				state, recoveryErr := configure.GetHostState()
				if recoveryErr == nil && state != nil {
					recoveryErr = RecoverHostState(state)
					if recoveryErr == nil {
						recoveryErr = configure.SetHostState(nil)
					}
				}
				err = errors.Join(err, recoveryErr)
			}
		}
	}()
	if err != nil {
		return err
	}

	// Phase 3 (post-lock): heavy operations — transparent proxy setup (DNS,
	// TUN routes), connectivity monitor, and post-start hook.
	if err = m.afterStart(process, generation); err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if !m.ownsProcessLocked(process, generation) {
		return fmt.Errorf("core process exited before startup committed")
	}
	if err = configure.SetRunning(true); err != nil {
		return err
	}
	if err = configure.SetLastKernelExitStatus(configure.LastKernelExitRunning); err != nil {
		return err
	}
	if err = configure.SetHostState(nil); err != nil {
		return err
	}
	if !testing {
		ApiFeed.ProductMessage("running_state", map[string]interface{}{"running": true, "networkPaused": false})
	}
	return nil
}

// Running reports if v2ray-core is running.
func (m *CoreProcessManager) Running() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.p != nil
}

func (m *CoreProcessManager) Process() *Process {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.p
}
