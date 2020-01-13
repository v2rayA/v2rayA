package service

import (
	"V2RayA/persistence/configure"
	"V2RayA/tools/jwt"
	"errors"
	"time"
)

func Login(username, password string) (token string, err error) {
	if !IsValidAccount(username, password) {
		return "", errors.New("用户名或密码错误")
	}
	dur := 3 * time.Hour
	return jwt.MakeJWT(map[string]string{
		"uname": username,
	}, &dur)
}

func IsValidAccount(username, password string) bool {
	pwd, err := configure.GetPasswordOfAccount(username)
	if err != nil {
		return false
	}
	return pwd == jwt.CryptoPwd(password)
}

func Register(username, password string) (token string, err error) {
	if configure.ExistsAccount(username) {
		return "", errors.New("用户名已存在")
	}
	err = configure.SetAccount(username, jwt.CryptoPwd(password))
	if err != nil {
		return
	}
	return Login(username,password)
}

func ValidPasswordLength(password string) bool {
	return len(password) >= 5 && len(password) <= 32
}
