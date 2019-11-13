package handlers

import (
	"V2RayA/extra/copyfile"
	"V2RayA/extra/quickdown"
	"V2RayA/global"
	"V2RayA/models/touch"
	"V2RayA/tools"
	"errors"
	"github.com/gin-gonic/gin"
	gonanoid "github.com/matoous/go-nanoid"
	"log"
	"os"
	"strings"
)

func PutGFWList(ctx *gin.Context) {
	c, err := tools.GetHttpClientAutomatically()
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}

	if _, err := os.Stat(global.V2RAY_LOCATION_ASSET + "/h2y.dat"); err == nil {
		//本地文件存在，检查本地版本是否比远端还新
		t, err := tools.GetFileModTime(global.V2RAY_LOCATION_ASSET + "/h2y.dat")
		if err != nil {
			tools.ResponseError(ctx, err)
			return
		}
		tRemote, err := tools.GetRemoteGFWListUpdateTime(c)
		if err != nil {
			tools.ResponseError(ctx, err)
			return
		}
		if t.After(tRemote) {
			//那确实新，不更新了
			tools.ResponseError(ctx, errors.New(
				"目前最新版本为"+tRemote.Format("2006-01-02")+"，您的本地文件已最新，无需更新",
			))
			return
		}
	}

	/* 更新h2y.dat */
	id, _ := gonanoid.Nanoid()
	quickdown.SetHttpClient(c)
	i := 0
	for {
		err = quickdown.DownloadWithWorkersTo("https://github.com/ToutyRater/V2Ray-SiteDAT/raw/master/geofiles/h2y.dat", 10, "/tmp/"+id)
		if err != nil && i < 3 && strings.Contains(err.Error(), "head fail") {
			//建立连接问题，最多重试3次
			i++
			continue
		}
		break
	}
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	err = copyfile.CopyFile("/tmp/"+id, global.V2RAY_LOCATION_ASSET+"/h2y.dat")
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	err = os.Chmod(global.V2RAY_LOCATION_ASSET+"/h2y.dat", os.FileMode(0755))
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	t, err := tools.GetFileModTime(global.V2RAY_LOCATION_ASSET + "/h2y.dat")
	var localGFWListVersion string
	if err == nil {
		localGFWListVersion = t.Format("2006-01-02")
	}
	tools.ResponseSuccess(ctx, gin.H{
		"localGFWListVersion": localGFWListVersion,
	})
}

func PutSubscription(ctx *gin.Context) {
	var data touch.WhichTouch
	err := ctx.ShouldBindJSON(&data)
	tr := global.GetTouchRaw()
	tr.Lock() //写操作加锁
	defer tr.Unlock()
	index := data.ID - 1
	if err != nil || data.TYPE != touch.SubscriptionType || index < 0 || index >= len(tr.Subscriptions) {
		tools.ResponseError(ctx, errors.New("参数有误"))
		log.Println(err, data.TYPE, index, len(tr.Subscriptions))
		return
	}
	addr := tr.Subscriptions[index].Address
	c, err := tools.GetHttpClientAutomatically()
	if err != nil {
		reason := "尝试使用代理失败，建议修改设置为直连模式再试"
		tools.ResponseError(ctx, errors.New(reason))
		return
	}
	infos, err := tools.ResolveSubscriptionWithClient(addr, c)
	if err != nil {
		reason := "解析订阅地址失败: " + err.Error()
		log.Println(infos, err)
		tools.ResponseError(ctx, errors.New(reason))
		return
	}
	tsrs := make([]touch.TouchServerRaw, len(infos))
	var connectedServer *touch.TouchServerRaw
	if tr.ConnectedServer != nil {
		connectedServer, _ = tr.LocateServer(tr.ConnectedServer)
	}
	//将列表更换为新的，并且找到一个跟现在连接的server值相等的，设为Connected，如果没有，则断开连接
	finishFindConnected := false
	for i, info := range infos {
		tsr := touch.TouchServerRaw{
			VmessInfo: info.VmessInfo,
			Connected: false,
		}
		if !finishFindConnected && connectedServer != nil && connectedServer.VmessInfo == tsr.VmessInfo {
			tsr.Connected = true
			tr.ConnectedServer = &touch.WhichTouch{
				TYPE:        touch.SubscriptionServerType,
				ID:          i + 1,
				Sub:         index,
				PingLatency: nil,
			}
			finishFindConnected = true
		}
		tsrs[i] = tsr
	}
	if !finishFindConnected {
		err = disconnect(&tr)
		if err != nil {
			tools.ResponseError(ctx, errors.New("现连接的服务器已被更新且不包含在新的订阅中，在试图与其断开的过程中遇到失败"))
			return
		}
	}
	tr.Subscriptions[index].Servers = tsrs
	tr.Subscriptions[index].Status = touch.NewUpdateStatus()
	err = tr.WriteToFile()
	if err != nil {
		tools.ResponseError(ctx, err)
	}
	global.SetTouchRaw(&tr)
	GetTouch(ctx)
}
