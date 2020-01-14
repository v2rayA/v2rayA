package jwt

import (
	"V2RayA/tools"
	"encoding/base64"
	"fmt"
	"github.com/dgrijalva/jwt-go"
	"github.com/dgrijalva/jwt-go/request"
	"github.com/gin-gonic/gin"
	"github.com/json-iterator/go"
	"github.com/pkg/errors"
	"strconv"
	"strings"
	"time"
)

func JWTAuth(Admin bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := request.ParseFromRequest(ctx.Request, request.AuthorizationHeaderExtractor,
			func(token *jwt.Token) (interface{}, error) {
				// 我们使用固定的secret，直接返回就好
				return []byte(secret), nil
			})
		if err != nil {
			tools.Response(ctx, tools.UNAUTHORIZED, err.Error())
			ctx.Abort()
			return
		}
		//如果需要Admin权限
		mapClaims := token.Claims.(jwt.MapClaims)
		if Admin && mapClaims["admin"] == false {
			tools.ResponseError(ctx, errors.New("admin required"))
			ctx.Abort()
			return
		}
		//将用户名丢入参数
		ctx.Set("Name", mapClaims["name"])
		//在ctx.Next()前的都是before request，之后的是after request
		ctx.Next()
	}
}

func ValidToken(token, secret string) (err error) {
	arr := strings.Split(token, ".")
	if len(arr) != 3 {
		return errors.New("token核对失败, token格式有误")
	}
	sign := base64.RawURLEncoding.EncodeToString(tools.HMACSHA256(arr[0]+"."+arr[1], []byte(secret)))
	if sign != arr[2] { //签名核对失败
		return errors.New("token核对失败, 无效签名")
	}
	pl, err := GetJWTPayload(token)
	if err != nil { //decode发生意外
		return errors.New("token核对失败, 解析失败")
	}
	strExp, ok := pl["exp"]
	if ok {
		iExp, err := strconv.Atoi(strExp)
		if err != nil { //str转int失败
			return errors.New("token核对失败")
		}
		if time.Now().After(time.Unix(int64(iExp), 0)) { //token过期
			return errors.New("token核对失败, token过期")
		}
	}
	return nil
}

func MakeJWT(payload map[string]string, expDuration *time.Duration) (jwt string, err error) {
	headerJSON, _ := jsoniter.Marshal(map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	})
	if expDuration != nil {
		payload["exp"] = fmt.Sprint(time.Now().Add(*expDuration).Unix())
	}
	payloadJSON, err := jsoniter.Marshal(payload)
	if err != nil {
		return
	}
	bh := base64.RawURLEncoding.EncodeToString(headerJSON)
	bp := base64.RawURLEncoding.EncodeToString(payloadJSON)
	signBefore := bh + "." + bp
	signature := tools.HMACSHA256(signBefore, []byte(secret))
	bs := base64.RawURLEncoding.EncodeToString(signature)
	return signBefore + "." + bs, nil
}

func GetJWTPayload(jwt string) (payload map[string]string, err error) {
	arr := strings.Split(jwt, ".")
	pl, err := base64.RawURLEncoding.DecodeString(arr[1])
	if err != nil {
		return
	}
	err = jsoniter.Unmarshal(pl, &payload)
	return
}
