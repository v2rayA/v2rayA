package jwt

import (
	"github.com/dgrijalva/jwt-go/v4"
	"github.com/dgrijalva/jwt-go/v4/request"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"github.com/v2rayA/v2rayA/common"
	"strings"
	"time"
)

// stripBearerPrefixFromTokenString strips 'Bearer ' prefix from bearer token string
func stripBearerPrefixFromTokenString(tok string) (string, error) {
	// Should be a bearer token
	if len(tok) > 6 && strings.ToUpper(tok[0:7]) == "BEARER " {
		return tok[7:], nil
	}
	return tok, nil
}

// AuthorizationArgumentExtractor extracts bearer token from Argument header
// Uses PostExtractionFilter to strip "Bearer " prefix from header
var AuthorizationArgumentExtractor = &request.PostExtractionFilter{
	request.ArgumentExtractor{"Authorization"},
	stripBearerPrefixFromTokenString,
}

func JWTAuth(Admin bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := request.ParseFromRequest(ctx.Request, request.AuthorizationHeaderExtractor,
			func(token *jwt.Token) (interface{}, error) {
				return getSecret(), nil
			})
		if err != nil {
			if errors.Is(err, request.ErrNoTokenInRequest) {
				token, err = request.ParseFromRequest(ctx.Request, AuthorizationArgumentExtractor,
					func(token *jwt.Token) (interface{}, error) {
						return getSecret(), nil
					})
			}
			if err != nil {
				common.Response(ctx, common.UNAUTHORIZED, err.Error())
				ctx.Abort()
				return
			}
		}
		mapClaims := token.Claims.(jwt.MapClaims)
		exp, ok := mapClaims["exp"]
		if ok {
			fExp, ok := exp.(float64)
			if !ok {
				common.ResponseError(ctx, errors.New("bad token: 4"))
				ctx.Abort()
				return
			}
			if time.Now().After(time.Unix(int64(fExp), 0)) {
				common.ResponseError(ctx, errors.New("expired token"))
				ctx.Abort()
				return
			}
		}
		//如果需要Admin权限
		if Admin && mapClaims["admin"] == false {
			common.ResponseError(ctx, errors.New("admin required"))
			ctx.Abort()
			return
		}
		//将用户名丢入参数
		ctx.Set("Name", mapClaims["name"])
	}
}

func MakeJWT(payload map[string]string, expDuration *time.Duration) (string, error) {
	claims := jwt.MapClaims{}
	for k := range payload {
		claims[k] = payload[k]
	}
	if expDuration != nil {
		claims["exp"] = time.Now().Add(*expDuration).Unix()
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(getSecret())
}
