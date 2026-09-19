package v2ray

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/shirou/gopsutil/v3/mem"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/kernel/serverObj"
	"github.com/v2rayA/v2rayA/kernel/v2ray/asset"
	"github.com/v2rayA/v2rayA/kernel/v2ray/where"
	"github.com/v2rayA/v2rayA/pkg/util/log"
)

var NoConnectedServerErr = common.Coded("NO_SERVER_SELECTED", fmt.Errorf("no selected servers"), nil)

// Process is a v2ray-core process
type Process struct {
	// mutex protect the proc
	mutex          sync.Mutex
	proc           *os.Process
	procCancel     func() // cancel func for proc
	template       *Template
	tag2WhichIndex map[string]int
	done           chan struct{}
	expectedStop   atomic.Bool
}

func NewProcess(tmpl *Template,
	prestart func() error, poststart func() error,
	postUnexpectedStop func(p *Process),
) (*Process, error) {
	process := &Process{
		template: tmpl,
		done:     make(chan struct{}),
	}
	var err error
	var rollbackStage int
	defer func() {
		if err != nil {
			process.rollback(rollbackStage)
		}
	}()
	// The template's API producers were started with it; a start that
	// fails before the process owns them must stop them, or each failed
	// start leaves a goroutine polling a port nothing listens on.
	defer func() {
		if err != nil {
			_ = tmpl.Close()
		}
	}()

	// DNS 模块由 v2raya-core 进程内启动（基于 dns_module 配置段），
	// v2rayA 仅负责生成配置和在透明代理时应用防火墙规则。
	if tmpl.Setting != nil && tmpl.DnsModuleConfig != nil {
		log.Info("DNS module config written to v2raya-core config")
	}
	if tmpl.MultiObservatory != nil {
		// NOTICE: tag2WhichIndex is reliable because once connected servers are changed when v2ray is running,
		// the func UpdateV2RayConfig should be invoked and tag2WhichIndex will be regenerated.
		tag2WhichIndex := make(map[string]int)
		for i, tag := range tmpl.OutboundTags {
			tag2WhichIndex[tag] = i
		}
		process.tag2WhichIndex = tag2WhichIndex
	}
	err = WriteV2rayConfig(tmpl.ToConfigBytes())
	if err != nil {
		return nil, err
	}
	if err = tmpl.CheckInboundPortsOccupied(); err != nil {
		return nil, err
	}
	pCtx, cancel := context.WithCancel(context.Background())
	defer func() {
		if err != nil {
			cancel()
		}
	}()
	if tmpl.API == nil {
		log.Fatal("unexpected tmpl.API == nil")
	}
	process.procCancel = cancel
	if err = prestart(); err != nil {
		return nil, err
	}
	proc, err := StartCoreProcess(pCtx)
	if err != nil {
		var coded *common.CodedError
		if errors.As(err, &coded) {
			return nil, err
		}
		return nil, common.Coded("CORE_START_FAILED", err, map[string]interface{}{"detail": err.Error()})
	}
	rollbackStage = 1 // xray 已启动
	if err = poststart(); err != nil {
		return nil, err
	}
	process.proc = proc
	var unexpectedExiting atomic.Bool
	go func() {
		defer close(process.done)
		p, e := proc.Wait()
		if process.expectedStop.Load() {
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
		log.Warn("v2ray-core: %v", strings.Join(t, ": "))
		unexpectedExiting.Store(true)
	}()
	// ports to check
	portList := []string{strconv.Itoa(tmpl.ApiPort)}
	log.Trace("portList for connectivity test: %+v", portList)
	startTime := time.Now()
	startTimeOut := time.Duration(conf.GetEnvironmentConfig().CoreStartupTimeout) * time.Second
	for i := 0; i < len(portList); {
		conn, err := net.Dial("tcp", net.JoinHostPort("127.0.0.1", portList[i]))
		if err == nil {
			conn.Close()
			i++
			continue
		}
		if unexpectedExiting.Load() {
			return nil, common.Coded("CORE_START_FAILED", fmt.Errorf("v2raya_core exited right after starting; the reason is in the v2rayA log"), map[string]interface{}{"detail": "v2raya_core exited right after starting; the reason is in the v2rayA log"})
		}
		if time.Since(startTime) > startTimeOut {
			log.Info("Attempting to terminate timed-out process with SIGTERM")
			_ = proc.Signal(syscall.SIGTERM)
			err := fmt.Errorf("v2raya_core did not open its API port within %d s (--core-startup-timeout); the reason is in the v2rayA log", int(startTimeOut/time.Second))
			return nil, common.Coded("CORE_START_FAILED", err, map[string]interface{}{"detail": err.Error()})
		}
		time.Sleep(100 * time.Millisecond)
	}
	log.Trace("Cost of waiting for v2ray-core: %v", time.Since(startTime).String())
	return process, nil
}

type logInfoWriter struct {
}

func (w logInfoWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return 0, nil
	}
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
	// 停止 xray 进程（v2raya-core 内部会同时关闭 DNS 模块）
	if p.procCancel != nil {
		cancel := p.procCancel
		p.expectedStop.Store(true)
		p.procCancel = nil
		cancel()
		err := p.template.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Process) WaitUntilExit(ctx context.Context) error {
	if p == nil {
		return nil
	}
	if ctx == nil {
		<-p.done
		return nil
	}
	select {
	case <-p.done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
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
	v2rayBinPath, err := where.GetV2rayBinPath()
	if err != nil {
		return nil, err
	}
	dir := filepath.Dir(v2rayBinPath)
	var arguments = []string{
		v2rayBinPath,
		"run",
		"--config=" + asset.GetV2rayConfigPath(),
	}
	if confdir := asset.GetV2rayConfigDirPath(); confdir != "" {
		arguments = append(arguments, "--confdir="+confdir)
	}

	// Get asset directory
	assetDir := asset.GetV2rayLocationAssetOverride()
	// The core is told to look in assetDir and nowhere else, so put the dat
	// files there first when they live in a system directory.
	asset.EnsureCoreAssets(assetDir)
	log.Info("Asset directory for %s: %v", "v2raya_core", assetDir)

	// Prepare environment variables, filtering out duplicates
	env := make([]string, 0, len(os.Environ())+4)
	for _, e := range os.Environ() {
		// Skip existing V2RAY_LOCATION_ASSET, XRAY_LOCATION_ASSET, V2RAY_CONF_GEOLOADER
		if strings.HasPrefix(e, "V2RAY_LOCATION_ASSET=") ||
			strings.HasPrefix(e, "XRAY_LOCATION_ASSET=") ||
			strings.HasPrefix(e, "V2RAY_CONF_GEOLOADER=") {
			continue
		}
		env = append(env, e)
	}

	// Add asset directory to environment based on core type.
	// v2raya_core is based on xray-core and uses XRAY_LOCATION_ASSET.
	env = append(env, "XRAY_LOCATION_ASSET="+assetDir)

	// Check memory and set geoloader mode
	memstat, err := mem.VirtualMemory()
	if err != nil {
		log.Warn("cannot get memory info: %v", err)
	} else {
		if memMiB := memstat.Available / 1024 / 1024; memMiB < 2048 {
			env = append(env, "V2RAY_CONF_GEOLOADER=memconservative")
			log.Info("low memory: %vMiB, set V2RAY_CONF_GEOLOADER=memconservative", memMiB)
		}
	}

	log.Debug(strings.Join(arguments, " "))
	proc, err := RunWithLog(ctx, v2rayBinPath, arguments, dir, env)
	if err != nil {
		return nil, err
	}
	return proc, nil
}

func findAvailablePluginPorts(vms []serverObj.ServerObj) (pluginPortMap map[int]int, err error) {
	pluginPortMap = make(map[int]int)
	for i, v := range vms {
		if !v.NeedPluginPort() {
			continue
		}
		//find a port that not be occupied
		var port int
		for {
			// Find a port >= 30000
			r := 30000 + common.RandInt(35535)
			l, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%v", r))
			if err == nil {
				defer l.Close()
				port = l.Addr().(*net.TCPAddr).Port
				break
			}
			time.Sleep(30 * time.Millisecond)
		}
		pluginPortMap[i] = port
	}
	return pluginPortMap, nil
}

func getConnectedServerObjs() ([]serverObj.ServerObj, []serverInfo, error) {
	css := configure.GetConnectedServers()
	if css.Len() == 0 { //no connected server. stop v2ray-core.
		return nil, nil, nil
	}
	serverInfos := make([]serverInfo, 0, css.Len())
	loc := configure.NewLocator()
	for _, cs := range css.Get() {
		sr, err := loc.Locate(cs)
		if err != nil {
			return nil, nil, err
		}
		serverInfos = append(serverInfos, serverInfo{
			Info:         sr.ServerObj,
			OutboundName: cs.Outbound,
		})
	}
	serverInfos = applySelection(serverInfos, func(outbound string) string {
		return configure.GetOutboundSetting(outbound).Selected
	})
	serverObjs := make([]serverObj.ServerObj, 0, len(serverInfos))
	for _, info := range serverInfos {
		serverObjs = append(serverObjs, info.Info)
	}
	return serverObjs, serverInfos, nil
}

// applySelection keeps, for a group whose setting selects one member, only
// that member; a selection matching no member leaves the group balanced.
func applySelection(serverInfos []serverInfo, selectedOf func(outbound string) string) []serverInfo {
	selected := make(map[string]string)
	matched := make(map[string]bool)
	for _, info := range serverInfos {
		if _, ok := selected[info.OutboundName]; !ok {
			selected[info.OutboundName] = selectedOf(info.OutboundName)
		}
		link := selected[info.OutboundName]
		if link != "" && info.Info.ExportToURL() == link {
			matched[info.OutboundName] = true
		}
	}
	kept := serverInfos[:0]
	for _, info := range serverInfos {
		if matched[info.OutboundName] && info.Info.ExportToURL() != selected[info.OutboundName] {
			continue
		}
		kept = append(kept, info)
	}
	return kept
}

func NewTemplateFromConnectedServers(setting *configure.Setting) (tmpl *Template, err error) {
	//read the database and convert to the v2ray-core template
	serverObjs, serverInfos, err := getConnectedServerObjs()
	if err != nil {
		return nil, err
	}
	if len(serverObjs) == 0 {
		return nil, NoConnectedServerErr
	}
	var pluginPorts map[int]int
	if pluginPorts, err = findAvailablePluginPorts(serverObjs); err != nil {
		return nil, err
	}
	for i := range serverInfos {
		if port, ok := pluginPorts[i]; ok {
			serverInfos[i].PluginPort = port
		}
	}
	tmpl, err = NewTemplate(serverInfos, setting)
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

func UpdateV2RayConfig() (err error) {
	tmpl, err := NewTemplateFromConnectedServers(nil)
	if err != nil {
		if errors.Is(err, NoConnectedServerErr) {
			//no servers are selected, which means to stop the v2ray-core
			ProcessManager.Stop(true)
			return nil
		}
		return err
	}
	err = ProcessManager.Start(tmpl)
	if err != nil {
		return err
	}
	return
}

// rollback 回滚已启动的组件。
// 注意：透明代理规则回滚由 processManager.afterStart 的 defer 处理。
func (p *Process) rollback(stage int) {
	if stage == 1 {
		// xray 已启动 → 回滚
		if p.procCancel != nil {
			p.procCancel()
			p.procCancel = nil
		}
	}
}
