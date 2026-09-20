package controller

import (
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
	v2ray "github.com/v2rayA/v2rayA/kernel/v2ray"
	"github.com/v2rayA/v2rayA/kernel/v2ray/asset/dat"
	"github.com/v2rayA/v2rayA/kernel/v2ray/service"
	"github.com/v2rayA/v2rayA/kernel/v2ray/where"
	"github.com/v2rayA/v2rayA/pkg/util/privilege"
)

func GetVersion(ctx *gin.Context) {
	var lite int
	if conf.GetEnvironmentConfig().Lite {
		lite = 1
	}

	// Detect if running as root (Linux/macOS only)
	isRoot := false
	switch runtime.GOOS {
	case "linux", "darwin":
		isRoot = os.Geteuid() == 0
	case "windows":
		isRoot = privilege.IsRootOrAdmin()
	}

	versionErr := service.CheckCoreVersionMatch()
	// the core's own version, for the dashboard; empty when the binary cannot be asked
	coreVersion := ""
	if _, ver, err := where.GetV2rayServiceVersion(); err == nil {
		coreVersion = ver
	}

	common.ResponseSuccess(ctx, gin.H{
		"version":          conf.Version,
		"foundNew":         conf.FoundNew,
		"remoteVersion":    conf.RemoteVersion,
		"serviceValid":     service.IsV2rayServiceValid(),
		"v5":               versionErr == nil,
		"lite":             lite,
		"loadBalanceValid": true,
		"variant":          where.V2rayaCore,
		"coreVersion":      coreVersion,
		"os":               runtime.GOOS,
		"isRoot":           isRoot,
		"tunSupported":     v2ray.TunSupported(),
		"coreVersionValid": versionErr == nil,
		"coreVersionErr": func() string {
			if versionErr != nil {
				return versionErr.Error()
			}
			return ""
		}(),
		"hasAccounts":    configure.HasAnyAccounts(),
		"lastKernelExit": configure.GetLastKernelExitStatus(),
		"docker":         common.IsDocker(),
	})
}

func GetRemoteGFWListVersion(ctx *gin.Context) {
	g, err := dat.GetRemoteGFWListUpdateTime(&http.Client{Timeout: 10 * time.Second})
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}
	common.ResponseSuccess(ctx, gin.H{"remoteGFWListVersion": g.UpdateTime.Local().Format("2006-01-02")})
}
