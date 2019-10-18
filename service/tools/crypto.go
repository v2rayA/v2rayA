package tools

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"io"
)

//用一个花里胡哨的加密方式来加密密码，密码最多32位
func CryptoPwd(password string) string {
	shaed := sha512.Sum512_256([]byte(password))
	pwd := md5.Sum(shaed[:])
	return fmt.Sprintf("%x", pwd)
}

// HMACSHA256
func HMACSHA256(s string, secret []byte) []byte {
	h := hmac.New(sha256.New, secret)
	io.WriteString(h, s)
	return h.Sum(nil)
}
