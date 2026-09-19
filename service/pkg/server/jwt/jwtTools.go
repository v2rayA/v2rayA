package jwt

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/golang-jwt/jwt/v5/request"
	"github.com/pkg/errors"
	"github.com/v2rayA/v2rayA/common"
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
	Extractor: request.ArgumentExtractor{"Authorization"},
	Filter:    stripBearerPrefixFromTokenString,
}

func JWTAuth(Admin bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		parser := jwt.NewParser(jwt.WithValidMethods([]string{"HS256"}))
		token, err := request.ParseFromRequest(ctx.Request, request.AuthorizationHeaderExtractor,
			func(token *jwt.Token) (interface{}, error) {
				return getSecret(), nil
			},
			request.WithParser(parser),
		)
		if err != nil {
			if errors.Is(err, request.ErrNoTokenInRequest) {
				token, err = request.ParseFromRequest(ctx.Request, AuthorizationArgumentExtractor,
					func(token *jwt.Token) (interface{}, error) {
						return getSecret(), nil
					},
					request.WithParser(parser),
				)
			}
			if err != nil {
				code := "SESSION_INVALID"
				if errors.Is(err, jwt.ErrTokenExpired) {
					code = "SESSION_EXPIRED"
				}
				common.Response(ctx, common.UNAUTHORIZED, common.Coded(code, err, nil))
				ctx.Abort()
				return
			}
		}
		mapClaims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			common.ResponseError(ctx, common.Coded("SESSION_INVALID", errors.New("the session token is not valid; sign in again"), nil))
			ctx.Abort()
			return
		}
		exp, err := mapClaims.GetExpirationTime()
		if err == nil && exp != nil {
			if time.Now().After(exp.Time) {
				common.ResponseError(ctx, common.Coded("SESSION_EXPIRED", errors.New("session expired; sign in again"), nil))
				ctx.Abort()
				return
			}
		}
		//如果需要Admin权限
		if Admin {
			adminVal, _ := mapClaims["admin"]
			if adminVal != true {
				common.ResponseError(ctx, common.Coded("ADMIN_REQUIRED", errors.New("this action needs an admin account"), nil))
				ctx.Abort()
				return
			}
		}
		//将用户名丢入参数
		if name, ok := mapClaims["uname"]; ok && name != nil {
			ctx.Set("Name", name)
		}
	}
}

func MakeJWT(payload map[string]string, expDuration *time.Duration) (string, error) {
	claims := jwt.MapClaims{}
	for k := range payload {
		claims[k] = payload[k]
	}
	if expDuration != nil {
		claims["exp"] = jwt.NewNumericDate(time.Now().Add(*expDuration))
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(getSecret())
}

// ValidateToken validates a JWT token string and returns true if valid
func ValidateToken(tokenString string) bool {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return getSecret(), nil
	}, jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return false
	}
	return token.Valid
}
