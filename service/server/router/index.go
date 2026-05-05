package router

import (
	"crypto/md5"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/conf"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/server/jwt"
	"github.com/v2rayA/v2rayA/pkg/server/reqCache"
	"github.com/v2rayA/v2rayA/pkg/util/log"
	"github.com/v2rayA/v2rayA/server/controller"
	"github.com/vearutop/statigz"
)

//go:embed web
var webRoot embed.FS

// relativeFS implements fs.FS
type relativeFS struct {
	root        fs.FS
	relativeDir string
}

func (c relativeFS) Open(name string) (fs.File, error) {
	return c.root.Open(filepath.Join(c.relativeDir, name))
}

func cachedHTML(html []byte) func(ctx *gin.Context) {
	etag := fmt.Sprintf("W/%x", md5.Sum(html))
	h := string(html)
	return func(ctx *gin.Context) {
		if ctx.IsAborted() {
			return
		}
		ctx.Header("Content-Type", "text/html; charset=utf-8")
		ctx.Header("Cache-Control", "public, must-revalidate")
		ctx.Header("ETag", etag)
		if match := ctx.GetHeader("If-None-Match"); match != "" {
			if strings.Contains(match, etag) {
				ctx.Status(http.StatusNotModified)
				return
			}
		}
		ctx.String(http.StatusOK, h)
	}
}

func safeStatigzHandler(fsys fs.FS) (http.Handler, bool) {
	fi, err := fs.Stat(fsys, ".")
	if err != nil || !fi.IsDir() {
		return nil, false
	}

	readDirFS, ok := fsys.(fs.ReadDirFS)
	if !ok {
		return nil, false
	}

	var handler http.Handler
	panicked := false
	func() {
		defer func() {
			if recover() != nil {
				panicked = true
			}
		}()
		handler = statigz.FileServer(readDirFS)
	}()

	if panicked || handler == nil {
		return nil, false
	}

	return handler, true
}

func registerEmbeddedRoute(r *gin.Engine, routePrefix string, fsCandidates ...fs.FS) bool {
	for _, candidate := range fsCandidates {
		if candidate == nil {
			continue
		}

		handler, ok := safeStatigzHandler(candidate)
		if !ok {
			continue
		}

		stripped := http.StripPrefix(routePrefix, handler)
		r.GET(routePrefix+"/*w", func(c *gin.Context) {
			stripped.ServeHTTP(c.Writer, c.Request)
		})
		return true
	}

	return false
}

func ServeGUI(r *gin.Engine) {
	webDir := conf.GetEnvironmentConfig().WebDir
	if webDir == "" {
		webFS, err := fs.Sub(webRoot, "web")
		if err != nil {
			log.Fatal("fs.Sub: %v", err)
		}

		// --- /static/* route (used by legacy gui Vite build) ---
		var staticCandidates []fs.FS
		if sub, subErr := fs.Sub(webFS, "static"); subErr == nil {
			staticCandidates = append(staticCandidates, sub)
		}
		// Backward-compatible fallback for legacy builds that emit hashed assets at web root.
		staticCandidates = append(staticCandidates, webFS)
		_ = registerEmbeddedRoute(r, "/static", staticCandidates...)

		// --- /_nuxt/* route (used by ngui Nuxt 3 build) ---
		if sub, subErr := fs.Sub(webFS, "_nuxt"); subErr == nil {
			_ = registerEmbeddedRoute(r, "/_nuxt", sub)
		}

		f, err := webFS.Open("index.html")
		if err != nil {
			log.Fatal("webFS.Open index.html:", err)
		}
		defer f.Close()
		html, err := io.ReadAll(f)
		if err != nil {
			log.Fatal("ReadAll index.html: %v", err)
		}
		r.GET("/", cachedHTML(html))
		if favicon, favErr := webFS.Open("favicon.ico"); favErr == nil {
			defer favicon.Close()
			favData, _ := io.ReadAll(favicon)
			r.GET("/favicon.ico", func(c *gin.Context) {
				c.Data(http.StatusOK, "image/x-icon", favData)
			})
		}
	} else {
		if _, err := os.Stat(webDir); os.IsNotExist(err) {
			log.Warn("web files cannot be found at %v. web UI cannot be served", webDir)
		} else {
			// --- /static/* route (dev mode, legacy gui) ---
			staticDir := filepath.Join(webDir, "static")
			if info, statErr := os.Stat(staticDir); statErr == nil && info.IsDir() {
				r.Static("/static", staticDir)
			} else {
				// Backward-compatible fallback for legacy builds that emit hashed assets at web root.
				r.Static("/static", webDir)
			}

			// --- /_nuxt/* route (dev mode, ngui) ---
			// Directory not existing is silently skipped, compatible with legacy gui dev mode
			nuxtDir := filepath.Join(webDir, "_nuxt")
			if info, statErr := os.Stat(nuxtDir); statErr == nil && info.IsDir() {
				r.Static("/_nuxt", nuxtDir)
			}

			f, err := os.Open(filepath.Join(webDir, "index.html"))
			if err != nil {
				log.Fatal("Open index.html: %v", err)
			}
			defer f.Close()
			html, err := io.ReadAll(f)
			if err != nil {
				log.Fatal("ReadAll index.html: %v", err)
			}
			r.GET("/", cachedHTML(html))
			favPath := filepath.Join(webDir, "favicon.ico")
			if _, favErr := os.Stat(favPath); favErr == nil {
				r.StaticFile("/favicon.ico", favPath)
			}
		}
	}

	app := conf.GetEnvironmentConfig()

	ip, port, _ := net.SplitHostPort(app.Address)
	addrs, err := net.InterfaceAddrs()
	if net.ParseIP(ip).IsUnspecified() && err == nil {
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok {
				printRunningAt("http://" + net.JoinHostPort(ipnet.IP.String(), port))
			}
		}
	} else {
		printRunningAt("http://" + app.Address)
	}
}

func nocache(c *gin.Context) {
	c.Header("Cache-Control", "no-cache, no-store, max-age=0, must-revalidate")
	c.Header("Pragma", "no-cache")
	c.Header("Expires", "0")
}

func Run() error {
	engine := gin.New()
	//ginpprof.Wrap(engine)
	engine.Use(gin.Recovery())
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowAllOrigins = true
	corsConfig.AllowMethods = []string{
		"GET", "POST", "DELETE", "PUT", "PATCH", "OPTIONS", "HEAD",
	}
	corsConfig.AllowWebSockets = true
	corsConfig.AllowCredentials = true
	corsConfig.AddAllowHeaders("Authorization", common.RequestIdHeader)
	corsConfig.AddExposeHeaders("Content-Disposition")
	engine.Use(cors.New(corsConfig))
	noAuth := engine.Group("api",
		nocache,
		reqCache.ReqCache,
	)
	{
		noAuth.GET("version", controller.GetVersion)
		noAuth.POST("login", controller.PostLogin)
		noAuth.GET("account", controller.GetAccount)
		noAuth.POST("account", controller.PostAccount)
	}
	auth := engine.Group("api",
		nocache,
		func(ctx *gin.Context) {
			if !configure.HasAnyAccounts() {
				common.Response(ctx, common.UNAUTHORIZED, gin.H{
					"first": true,
				})
				ctx.Abort()
				return
			}
		},
		jwt.JWTAuth(false),
		reqCache.ReqCache,
	)
	{
		auth.POST("import", controller.PostImport)
		auth.GET("touch", controller.GetTouch)
		auth.DELETE("touch", controller.DeleteTouch)
		auth.POST("connection", controller.PostConnection)
		auth.DELETE("connection", controller.DeleteConnection)
		auth.POST("v2ray", controller.PostV2ray)
		auth.DELETE("v2ray", controller.DeleteV2ray)
		auth.GET("pingLatency", controller.GetPingLatency)
		auth.GET("httpLatency", controller.GetHttpLatency)
		auth.GET("sharingAddress", controller.GetSharingAddress)
		auth.GET("remoteGFWListVersion", controller.GetRemoteGFWListVersion)
		auth.GET("setting", controller.GetSetting)
		auth.PUT("setting", controller.PutSetting)
		auth.PUT("gfwList", controller.PutGFWList)
		auth.DELETE("gfwList", controller.DeleteGFWList)
		auth.PUT("subscription", controller.PutSubscription)
		auth.PATCH("subscription", controller.PatchSubscription)
		auth.GET("ports", controller.GetPorts)
		auth.PUT("ports", controller.PutPorts)
		auth.GET("customInbound", controller.GetCustomInbound)
		auth.POST("customInbound", controller.PostCustomInbound)
		auth.DELETE("customInbound", controller.DeleteCustomInbound)
		//auth.PUT("account", controller.PutAccount)
		auth.GET("dnsRules", controller.GetDnsRules)
		auth.PUT("dnsRules", controller.PutDnsRules)
		auth.GET("routingA", controller.GetRoutingA)
		auth.PUT("routingA", controller.PutRoutingA)
		auth.GET("outbounds", controller.GetOutbounds)
		auth.GET("outbound", controller.GetOutbound)
		auth.POST("outbound", controller.PostOutbound)
		auth.PUT("outbound", controller.PutOutbound)
		auth.PUT("outboundConnections", controller.PutOutboundConnections)
		auth.DELETE("outbound", controller.DeleteOutbound)
		auth.GET("message", controller.WsMessage)
		auth.GET("logger", controller.GetLogger)
		auth.GET("domainsExcluded", controller.GetDomainsExcluded)
		auth.GET("tproxyWhiteIpGroups", controller.GetTproxyWhiteIpGroups)
		auth.PUT("domainsExcluded", controller.PutDomainsExcluded)
		auth.PUT("tproxyWhiteIpGroups", controller.PutTproxyWhiteIpGroups)
		auth.GET("networkInterfaces", controller.GetNetworkInterfaces)
	}

	ServeGUI(engine)

	return engine.Run(conf.GetEnvironmentConfig().Address)
}

func printRunningAt(address string) {
	log.Alert("v2rayA is listening at %v", address)
}
