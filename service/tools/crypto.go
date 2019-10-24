package tools

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os/exec"
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

// 封装一个外部调用进行base64解码
func ExecBase64Decode(s string) (string, error) {
	if len(s)%4 > 0 {
		s += strings.Repeat("=", 4-len(s)%4)
	}
	rbuf := new(bytes.Buffer)
	ebuf := new(bytes.Buffer)
	c1 := exec.Command("echo", s)
	c2 := exec.Command("base64", "-d")
	c2.Stdin, _ = c1.StdoutPipe()
	c2.Stdout = rbuf
	c2.Stderr = ebuf
	c2.Start()
	c1.Run()
	c2.Wait()
	var err error
	if ebuf.Len() > 0 {
		err = errors.New(ebuf.String())
	}
	return rbuf.String(), err
}
