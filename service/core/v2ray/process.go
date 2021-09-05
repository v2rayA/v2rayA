package v2ray

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/v2rayA/v2rayA/core/serverObj"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"github.com/v2rayA/v2rayA/core/v2ray/where"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"net"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"
)

// Process is a v2ray-core process
type Process struct {
	proc           *os.Process
	template       *Template
	tag2WhichIndex map[string]int
}

func NewProcess(tmpl *Template) (process *Process, err error) {
	process = &Process{
		template: tmpl,
	}
	if tmpl.Observatory != nil {
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
		return nil, fmt.Errorf("%v", err)
	}
	go tmpl.ServePlugins()
	defer func() {
		if err != nil {
			_ = tmpl.Close()
		}
	}()
	if tmpl.API == nil {
		tmpl.SetAPI()
	}
	proc, err := StartCoreProcess()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			e := proc.Kill()
			if e == nil {
				proc.Wait()
			}
		}
	}()
	startTime := time.Now()
	for {
		conn, err := net.Dial("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(tmpl.ApiPort)))
		if err == nil {
			conn.Close()
			break
		}

		if time.Since(startTime) > 15*time.Second {
			return nil, fmt.Errorf("v2ray-core does not start normally, check the log for more information")
		}
		time.Sleep(1000 * time.Millisecond)
	}
	process.proc = proc
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
	log.Info("%v", s)
	return len(p), nil
}

var logWriter logInfoWriter

func (p *Process) Close() error {
	err := p.proc.Kill()
	if err != nil {
		return err
	}
	p.proc.Wait()
	err = p.template.Close()
	if err != nil {
		return err
	}
	return nil
}

func RunWithLog(name string, argv []string, dir string, env []string) (*os.Process, error) {
	cmd := exec.Command(name)
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

func StartCoreProcess() (*os.Process, error) {
	v2rayBinPath, err := where.GetV2rayBinPath()
	if err != nil {
		return nil, err
	}
	dir := path.Dir(v2rayBinPath)
	var arguments = []string{
		v2rayBinPath,
		"--config=" + asset.GetV2rayConfigPath(),
	}
	if confdir := asset.GetV2rayConfigDirPath(); confdir != "" {
		arguments = append(arguments, "--confdir="+confdir)
	}
	log.Debug(strings.Join(arguments, " "))
	assetDir := asset.GetV2rayLocationAsset()
	env := append(os.Environ(),
		"V2RAY_LOCATION_ASSET="+assetDir,
		"XRAY_LOCATION_ASSET="+assetDir,
	)
	if CheckMemconservativeSupported() == nil {
		memstat, err := mem.VirtualMemory()
		if err != nil {
			log.Warn("cannot get memory info: %v", err)
		} else {
			if memMiB := memstat.Available / 1024 / 1024; memMiB < 2048 {
				env = append(env, "V2RAY_CONF_GEOLOADER=memconservative")
				log.Info("low memory: %vMiB, set V2RAY_CONF_GEOLOADER=memconservative", memMiB)
			}
		}
	}
	proc, err := RunWithLog(v2rayBinPath, arguments, dir, env)
	if err != nil {
		return nil, err
	}
	return proc, nil
}

func findAvailablePluginPorts(vms []serverObj.ServerObj) (pluginPortMap map[int]int, err error) {
	pluginPortMap = make(map[int]int)
	for i, v := range vms {
		if !v.NeedPlugin() {
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
			time.Sleep(100 * time.Millisecond)
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

func UpdateV2RayConfig() (err error) {
	//read the database and convert to the v2ray-core template
	serverObjs, serverInfos, err := getConnectedServerObjs()
	if err != nil {
		return err
	}
	if len(serverObjs) == 0 {
		//no servers are selected, which means to stop the v2ray-core
		ProcessManager.Stop(true)
		return nil
	}
	var pluginPorts map[int]int
	if pluginPorts, err = findAvailablePluginPorts(serverObjs); err != nil {
		return err
	}
	for i := range serverInfos {
		if port, ok := pluginPorts[i]; ok {
			serverInfos[i].PluginPort = port
		}
	}
	tmpl, err := NewTemplate(serverInfos)
	if err != nil {
		return
	}
	err = ProcessManager.Start(tmpl)
	if err != nil {
		return err
	}
	return
}
