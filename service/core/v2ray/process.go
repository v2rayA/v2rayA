package v2ray

import (
	"context"
	"errors"
	"fmt"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/core/serverObj"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"github.com/v2rayA/v2rayA/core/v2ray/where"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

var NoConnectedServerErr = fmt.Errorf("no selected servers")

// Process is a v2ray-core process
type Process struct {
	// mutex protect the proc
	mutex          sync.Mutex
	proc           *os.Process
	procCancel     func() // cancel func for proc and pluginManagers
	pluginManagers []*os.Process
	template       *Template
	tag2WhichIndex map[string]int
}

func NewProcess(tmpl *Template,
	prestart func() error, poststart func() error,
	postUnexpectedStop func(p *Process),
) (*Process, error) {
	process := &Process{
		template: tmpl,
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
	err := WriteV2rayConfig(tmpl.ToConfigBytes())
	if err != nil {
		return nil, err
	}
	if err = tmpl.CheckInboundPortsOccupied(); err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	go tmpl.ServePlugins()
	pCtx, cancel := context.WithCancel(context.Background())
	defer func() {
		if err != nil {
			cancel()
		}
	}()
	// start PluginManagers
	if pm := conf.GetEnvironmentConfig().PluginManager; pm != "" {
		for _, v := range tmpl.PluginManagerInfoList {
			arguments := []string{
				pm,
				"--stage=run",
				fmt.Sprintf("--link=%v", v.Link),
				fmt.Sprintf("--port=%v", v.Port),
				fmt.Sprintf("--v2raya-confdir=%v", conf.GetEnvironmentConfig().Config),
			}
			proc, err := RunWithLog(pCtx, pm, arguments, "", os.Environ())
			if err != nil {
				// clean
				for _, pm := range process.pluginManagers {
					_ = pm.Kill()
				}
				process.pluginManagers = nil
				return nil, fmt.Errorf("executing PluginManager [state: run, link: %v]: %w", v.Link, err)
			}
			process.pluginManagers = append(process.pluginManagers, proc)
		}
	}
	defer func() {
		if err != nil {
			_ = tmpl.Close()
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
		return nil, err
	}
	if err = poststart(); err != nil {
		return nil, err
	}
	process.proc = proc
	var unexpectedExiting bool
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
		log.Warn("v2ray-core: %v", strings.Join(t, ": "))
		unexpectedExiting = true
	}()
	// ports to check
	portList := []string{strconv.Itoa(tmpl.ApiPort)}
	for _, plu := range tmpl.Plugins {
		_, port, err := net.SplitHostPort(plu.ListenAddr())
		if err != nil {
			return nil, err
		}
		portList = append(portList, port)
	}
	log.Trace("portList for connectivity test: %+v", portList)
	startTime := time.Now()
	for i := 0; i < len(portList); {
		conn, err := net.Dial("tcp", net.JoinHostPort("127.0.0.1", portList[i]))
		if err == nil {
			conn.Close()
			i++
			continue
		}
		if unexpectedExiting {
			if log.Log.GetLevel() > log.ParseLevel("info") {
				log.Error("some critical information may lost due to your log level")
			}
			return nil, fmt.Errorf("unexpected exiting: check the log for more information")
		}
		if time.Since(startTime) > 15*time.Second {
			return nil, fmt.Errorf("timeout: check the log for more information")
		}
		time.Sleep(100 * time.Millisecond)
	}
	log.Trace("Cost of waiting for v2ray-core: %v", time.Since(startTime).String())
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
		err := p.template.Close()
		if err != nil {
			return err
		}
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
	log.Debug(strings.Join(arguments, " "))
	assetDir := asset.GetV2rayLocationAssetOverride()
	env := append(
		os.Environ(),
		"V2RAY_LOCATION_ASSET="+assetDir,
		"XRAY_LOCATION_ASSET="+assetDir,
	)
	memstat, err := mem.VirtualMemory()
	if err != nil {
		log.Warn("cannot get memory info: %v", err)
	} else {
		if memMiB := memstat.Available / 1024 / 1024; memMiB < 2048 {
			env = append(env, "V2RAY_CONF_GEOLOADER=memconservative")
			log.Info("low memory: %vMiB, set V2RAY_CONF_GEOLOADER=memconservative", memMiB)
		}
	}
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
			l, err := net.Listen("tcp", "127.0.0.1:0")
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
	serverObjs := make([]serverObj.ServerObj, 0, css.Len())
	for _, cs := range css.Get() {
		sr, err := cs.LocateServerRaw()
		if err != nil {
			return nil, nil, err
		}
		serverInfos = append(serverInfos, serverInfo{
			Info:         sr.ServerObj,
			OutboundName: cs.Outbound,
		})
		serverObjs = append(serverObjs, sr.ServerObj)
	}
	return serverObjs, serverInfos, nil
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
