package controller

import (
	"net"
	"runtime"

	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
)

type networkInterfaceInfo struct {
	Name       string   `json:"name"`
	Addrs      []string `json:"addrs"`
	IsLoopback bool     `json:"isLoopback"`
}

// GetNetworkInterfaces returns all active network interfaces with their addresses.
// On Windows the interface names are GUIDs / friendly descriptions, so the Name
// field carries the display-friendly value from net.Interface.Name.
func GetNetworkInterfaces(ctx *gin.Context) {
	ifaces, err := net.Interfaces()
	if err != nil {
		common.ResponseError(ctx, logError(err))
		return
	}

	var result []networkInterfaceInfo
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		var addrStrings []string
		for _, addr := range addrs {
			addrStrings = append(addrStrings, addr.String())
		}
		_ = runtime.GOOS // imported for potential future platform-specific logic
		result = append(result, networkInterfaceInfo{
			Name:       iface.Name,
			Addrs:      addrStrings,
			IsLoopback: iface.Flags&net.FlagLoopback != 0,
		})
	}
	common.ResponseSuccess(ctx, gin.H{"interfaces": result})
}
