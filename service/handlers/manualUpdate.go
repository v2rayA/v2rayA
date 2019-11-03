package handlers

import (
	"V2RayA/extra/copyfile"
	"V2RayA/extra/quickdown"
	"V2RayA/tools"
	"errors"
	"github.com/gin-gonic/gin"
	gonanoid "github.com/matoous/go-nanoid"
	"os"
	"strings"
)

func PutGFWList(ctx *gin.Context) {
	c, err := tools.GetHttpClientAutomatically()
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}

	if _, err := os.Stat("/etc/v2ray/h2y.dat"); err == nil {
		//本地文件存在，检查本地版本是否比远端还新
		t, err := tools.GetFileModTime("/etc/v2ray/h2y.dat")
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

	/* 更新/etc/v2ray/h2y.dat */
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
	err = copyfile.CopyFile("/tmp/"+id, "/etc/v2ray/h2y.dat")
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	err = os.Chmod("/etc/v2ray/h2y.dat", os.FileMode(0755))
	if err != nil {
		tools.ResponseError(ctx, err)
		return
	}
	t, err := tools.GetFileModTime("/etc/v2ray/h2y.dat")
	var localGFWListVersion string
	if err == nil {
		localGFWListVersion = t.Format("2006-01-02")
	}
	tools.ResponseSuccess(ctx, gin.H{
		"localGFWListVersion": localGFWListVersion,
	})
}
