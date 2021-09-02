package v2ray

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/json-iterator/go"
	"github.com/v2rayA/v2rayA/common/ntp"
	"github.com/v2rayA/v2rayA/common/resolv"
	"github.com/v2rayA/v2rayA/core/specialMode"
	"github.com/v2rayA/v2rayA/core/v2ray/asset"
	"github.com/v2rayA/v2rayA/core/v2ray/where"
	"github.com/v2rayA/v2rayA/core/vmessInfo"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/plugin"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"io"
	"net"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
)

var v2RayPID *Process
var tag2WhichIndex map[string]int
var apiPort int
var apiCloseFuncs []func()

func SetCoreProcess(p *Process) {
	configure.SetRunning(p != nil)
	v2RayPID = p
}
func CoreProcess() *Process {
	return v2RayPID
}
func IsV2RayRunning() bool {
	return CoreProcess() != nil
}
func ApiPort() int {
	if !IsV2RayRunning() {
		log.Trace("v2ray not running")
		return 0
	}
	return apiPort
}

// Process is a v2ray-core process
type Process struct {
	p       *os.Process
	logFile *os.File
	l       net.Listener
}

func (p *Process) LoggerBackground() {
	r := p.logFile
	go func() {
		p.p.Wait()
		p.logFile.Close()
	}()
	var buf [1024]byte
	var buff bytes.Buffer
	var off int64
	for {
		time.Sleep(500 * time.Millisecond)
		n, err := r.ReadAt(buf[:], off)
		if err != nil {
			if errors.Is(err, io.EOF) {
				if n > 0 {
					off += int64(n)
					var toWrite []byte
					if buff.Len() > 0 {
						buff.Write(buf[:n])
						toWrite = buff.Bytes()
					} else {
						toWrite = buf[:n]
					}
					lines := bytes.Split(toWrite, []byte("\n"))
					for _, line := range lines {
						if len(line) == 0 {
							continue
						}
						log.Info("%v", string(line))
					}
					if buff.Len() > 0 {
						buff.Reset()
					}
				}
				continue
			}
			log.Debug("LoggerBackground: %v", err)
			return
		}
		buff.Write(buf[:n])
	}
}

func (p *Process) Close() error {
	err := p.p.Kill()
	if err != nil {
		return err
	}
	p.p.Wait()
	return nil
}

func NewProcess(name string, argv []string, dir string, env []string) (*Process, error) {
	f, err := os.CreateTemp("", "v2raya_v2ray_*.log")
	if err != nil {
		log.Warn("cannot redirect v2ray log: %v", err)
		f = nil
	}
	p, err := os.StartProcess(name, argv, &os.ProcAttr{
		Dir: dir, //防止找不到v2ctl
		Env: env,
		Files: []*os.File{
			nil,
			f,
			f,
		},
	})
	if err != nil {
		return nil, err
	}
	proc := &Process{
		p:       p,
		logFile: f,
		l:       nil,
	}
	if f != nil {
		go proc.LoggerBackground()
	}
	return proc, nil
}

func StartCoreProcess() (*Process, error) {
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
	proc, err := NewProcess(v2rayBinPath, arguments, dir, append(os.Environ(),
		"V2RAY_LOCATION_ASSET="+assetDir,
		"XRAY_LOCATION_ASSET="+assetDir,
	))
	if err != nil {
		return nil, err
	}
	return proc, nil
}

func RestartV2rayService(saveStatus bool) (process *Process, err error) {
	if ok, t, err := ntp.IsDatetimeSynced(); err == nil && !ok {
		return nil, fmt.Errorf("Please sync datetime first. Your datetime is %v, and the correct datetime is %v", time.Now().Local().Format(ntp.DisplayFormat), t.Local().Format(ntp.DisplayFormat))
	}
	setting := configure.GetSettingNotNil()
	if (setting.Transparent == configure.TransparentGfwlist || setting.RulePortMode == configure.GfwlistMode) && !asset.IsGFWListExists() {
		return nil, fmt.Errorf("cannot find GFWList files. update GFWList and try again")
	}
	if err = killV2ray(saveStatus); err != nil {
		return
	}
	process, err = StartCoreProcess()
	if err != nil {
		return nil, fmt.Errorf("RestartV2rayService: %w", err)
	}
	if saveStatus {
		SetCoreProcess(process)
	}
	defer func() {
		if err != nil && CoreProcess() != nil {
			_ = killV2ray(true)
		}
	}()
	//如果inbounds中开放端口，检测端口是否已就绪
	tmplJson := Template{}
	var b []byte
	b, err = asset.GetConfigBytes()
	if err != nil {
		return
	}
	err = jsoniter.Unmarshal(b, &tmplJson)
	if err != nil {
		return
	}

	startTime := time.Now()
	for {
		conn, err := net.Dial("tcp", net.JoinHostPort("127.0.0.1", strconv.Itoa(apiPort)))
		if err == nil {
			conn.Close()
			break
		}

		if time.Since(startTime) > 15*time.Second {
			return nil, fmt.Errorf("v2ray-core does not start normally, check the log for more information")
		}
		time.Sleep(1000 * time.Millisecond)
	}
	return
}

func findAvailablePluginPorts(vms []vmessInfo.VmessInfo) (pluginPortMap map[int]int, err error) {
	pluginPortMap = make(map[int]int)
	for i, v := range vms {
		if !plugin.HasProperPlugin(v) {
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

func UpdateV2RayConfig() (err error) {
	resolv.CheckResolvConf()
	CheckAndStopTransparentProxy()
	defer func() {
		if err == nil {
			if e := CheckAndSetupTransparentProxy(true); e != nil {
				err = e
				if IsV2RayRunning() {
					if e = StopV2rayService(true); e != nil {
						err = fmt.Errorf("%w: %w", err, e)
					}
				}
			}
		}
		if err != nil {
			err = fmt.Errorf("UpdateV2RayConfig: %w", err)
		}
	}()
	plugin.GlobalPlugins.CloseAll()
	specialMode.StopDNSSupervisor()
	//read the database and convert to the v2ray-core template
	var (
		tmpl Template
		sr   *configure.ServerRaw
	)
	css := configure.GetConnectedServers()
	if css.Len() == 0 { //no connected server. stop v2ray-core.
		return StopV2rayService(true)
	}
	serverInfos := make([]serverInfo, 0, css.Len())
	vms := make([]vmessInfo.VmessInfo, 0, css.Len())
	for _, cs := range css.Get() {
		sr, err = cs.LocateServerRaw()
		if err != nil {
			return
		}
		serverInfos = append(serverInfos, serverInfo{
			Info:         sr.VmessInfo,
			OutboundName: cs.Outbound,
		})
		vms = append(vms, sr.VmessInfo)
	}
	var pluginPorts map[int]int
	if pluginPorts, err = findAvailablePluginPorts(vms); err != nil {
		return err
	}
	for i := range serverInfos {
		if port, ok := pluginPorts[i]; ok {
			serverInfos[i].PluginPort = port
		}
	}
	var outboundTags []string
	tmpl, outboundTags, err = NewTemplate(serverInfos)
	if err != nil {
		return
	}
	// NOTICE: tag2WhichIndex is reliable because once connected servers are changed when v2ray is running,
	// the func UpdateV2RayConfig should be invoked and tag2WhichIndex will be regenerated.
	tag2WhichIndex = make(map[string]int)
	for i, tag := range outboundTags {
		tag2WhichIndex[tag] = i
	}
	err = WriteV2rayConfig(tmpl.ToConfigBytes())
	if err != nil {
		return
	}

	if css.Len() == 0 && !IsV2RayRunning() {
		//no need to restart if no connected servers
		return
	}
	if err = tmpl.CheckInboundPortsOccupied(); err != nil {
		return fmt.Errorf("%v", err)
	}
	_, err = RestartV2rayService(true)
	if err != nil {
		return
	}

	// try launching plugin clients
	for i, v := range vms {
		var (
			port int
			ok   bool
		)
		if port, ok = pluginPorts[i]; !ok {
			continue
		}
		var plu plugin.Plugin
		plu, err = plugin.NewPluginAndServe(port, v)
		if err != nil {
			return
		}
		if plu != nil {
			plugin.GlobalPlugins.Add(css.Get()[i].Outbound, plu)
		}
	}

	specialMode.CheckAndSetupDNSSupervisor()
	return
}

func killV2ray(saveStatus bool) (err error) {
	if CoreProcess() != nil {
		err = CoreProcess().Close()
		if err != nil {
			if errors.Is(err, os.ErrProcessDone) {
				if saveStatus {
					SetCoreProcess(nil)
				}
			}
			return
		}
		if saveStatus {
			SetCoreProcess(nil)
		}
	}
	return
}

func StopV2rayService(saveStatus bool) (err error) {
	defer CheckAndStopTransparentProxy()
	defer plugin.GlobalPlugins.CloseAll()
	defer specialMode.StopDNSSupervisor()
	defer func() {
		if IsV2RayRunning() {
			msg := "failed to stop v2ray"
			if err != nil && len(strings.TrimSpace(err.Error())) > 0 {
				msg += ": " + err.Error()
			}
			err = fmt.Errorf("StopV2rayService: %v", msg)
		}
	}()
	return killV2ray(saveStatus)
}
