package service

import (
	"fmt"
	"github.com/v2rayA/v2rayA/common"
	"github.com/v2rayA/v2rayA/db/configure"
	"github.com/v2rayA/v2rayA/pkg/server/jwt"
	"time"
)

func Login(username, password string) (token string, err error) {
	if !IsValidAccount(username, password) {
		return "", fmt.Errorf("wrong username or password")
	}
	dur := 30 * 24 * time.Hour
	return jwt.MakeJWT(map[string]string{
		"uname": username,
	}, &dur)
}

func IsValidAccount(username, password string) bool {
	pwd, err := configure.GetPasswordOfAccount(username)
	if err != nil {
		return false
	}
	return pwd == common.CryptoPwd(password)
}

func Register(username, password string) (token string, err error) {
	if configure.ExistsAccount(username) {
		return "", fmt.Errorf("username exists")
	}
	err = configure.SetAccount(username, common.CryptoPwd(password))
	if err != nil {
		return
	}
	return Login(username, password)
}

func ValidPasswordLength(password string) (bool, error) {
	if len(password) >= 6 && len(password) <= 32 {
		return true, nil
	} else {
		return false, fmt.Errorf("length of password should be between 6 and 32")
	}
}
