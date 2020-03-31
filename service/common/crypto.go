package common

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

// 用一个花里胡哨的加密方式来加密密码，密码最多32位
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

// 封装base64.StdEncoding进行解码，加入了长度补全。当error时，返回输入和err
func Base64StdDecode(s string) (string, error) {
	s = strings.TrimSpace(s)
	saver := s
	if len(s)%4 > 0 {
		s += strings.Repeat("=", 4-len(s)%4)
	}
	raw, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return saver, err
	}
	return string(raw), err
}

// 封装base64.URLEncoding进行解码，加入了长度补全。当error时，返回输入和err
func Base64URLDecode(s string) (string, error) {
	s = strings.TrimSpace(s)
	saver := s
	if len(s)%4 > 0 {
		s += strings.Repeat("=", 4-len(s)%4)
	}
	raw, err := base64.URLEncoding.DecodeString(s)
	if err != nil {
		return saver, err
	}
	return string(raw), err
}