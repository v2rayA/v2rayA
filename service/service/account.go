package service

import (
	"V2RayA/common"
	"V2RayA/common/jwt"
	"V2RayA/persistence/configure"
	"time"
)

func Login(username, password string) (token string, err error) {
	if !IsValidAccount(username, password) {
		return "", newError("wrong username or password")
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
	return pwd == common.CryptoPwd(password)
}

func Register(username, password string) (token string, err error) {
	if configure.ExistsAccount(username) {
		return "", newError("username exists")
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
		return false, newError("length of password should be between 6 and 32")
	}
}
