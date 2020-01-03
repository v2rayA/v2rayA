module V2RayA

go 1.12

require (
	github.com/0xAX/notificator v0.0.0-20191016112426-3962a5ea8da1 // indirect
	github.com/Code-Hex/pget v0.0.0-20170428105109-9294f7465fa7
	github.com/Code-Hex/updater v0.0.0-20160712085121-c3f278672520 // indirect
	github.com/PuerkitoBio/goquery v1.5.0
	github.com/antonholmquist/jason v1.0.0 // indirect
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a // indirect
	github.com/cakturk/go-netstat v0.0.0-20190620190123-a633b9c55b1a
	github.com/codegangsta/envy v0.0.0-20141216192214-4b78388c8ce4 // indirect
	github.com/codegangsta/gin v0.0.0-20171026143024-cafe2ce98974 // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/dsnet/compress v0.0.1 // indirect
	github.com/fatih/color v1.7.0 // indirect
	github.com/frankban/quicktest v1.7.2 // indirect
	github.com/gin-contrib/cors v1.3.0
	github.com/gin-gonic/gin v1.4.0
	github.com/golang-collections/collections v0.0.0-20130729185459-604e922904d3 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/gookit/color v1.2.0
	github.com/jessevdk/go-flags v1.4.0 // indirect
	github.com/json-iterator/go v1.1.6
	github.com/matoous/go-nanoid v1.1.0
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-runewidth v0.0.5 // indirect
	github.com/mattn/go-shellwords v1.0.6 // indirect
	github.com/mcuadros/go-version v0.0.0-20190830083331-035f6764e8d2 // indirect
	github.com/mholt/archiver v3.1.1+incompatible // indirect
	github.com/mzz2017/shadowsocksR v0.0.0-20191227180500-1ac9689bb34d
	github.com/nadoo/glider v0.9.2
	github.com/nwaples/rardecode v1.0.0 // indirect
	github.com/pierrec/lz4 v2.4.0+incompatible // indirect
	github.com/pkg/errors v0.8.1
	github.com/ricochet2200/go-disk-usage v0.0.0-20150921141558-f0d1b743428f // indirect
	github.com/stevenroose/gonfig v0.1.4
	github.com/tidwall/gjson v1.3.5
	github.com/tidwall/sjson v1.0.4
	github.com/xi2/xz v0.0.0-20171230120015-48954b6210f8 // indirect
	golang.org/x/net v0.0.0-20191014212845-da9a3fd4c582
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e // indirect
	gopkg.in/cheggaaa/pb.v1 v1.0.28 // indirect
	gopkg.in/urfave/cli.v1 v1.20.0 // indirect
)

// Replace dependency modules with local developing copy
// use `go list -m all` to confirm the final module used
// replace github.com/mzz2017/shadowsocksR => ../../shadowsocksR
