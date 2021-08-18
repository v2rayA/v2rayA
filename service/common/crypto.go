package common

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"github.com/pkg/errors"
	"io"
	"math/big"
	"os"
	"strings"
	"time"
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

// StringToUUID5 is from https://github.com/XTLS/Xray-core/issues/158
func StringToUUID5(str string) string {
	var Nil [16]byte
	h := sha1.New()
	h.Write(Nil[:])
	h.Write([]byte(str))
	u := h.Sum(nil)[:16]
	u[6] = (u[6] & 0x0f) | (5 << 4)
	u[8] = u[8]&(0xff>>2) | (0x02 << 6)
	buf := make([]byte, 36)
	hex.Encode(buf[0:8], u[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], u[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], u[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], u[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:], u[10:])
	return string(buf)
}

func GetCertInfo(crt string) (sha []byte, commonName string, err error) {
	b, err := os.ReadFile(crt)
	if err != nil {
		return nil, "", err
	}
	p, _ := pem.Decode(b)
	if p == nil {
		return nil, "", fmt.Errorf("bad certificate")
	}
	cert, err := x509.ParseCertificate(p.Bytes)
	if err != nil {
		return nil, "", fmt.Errorf("bad certificate: %w", err)
	}
	pubDER, err := x509.MarshalPKIXPublicKey(cert.PublicKey.(*rsa.PublicKey))
	sum := sha256.Sum256(pubDER)
	pin := make([]byte, base64.StdEncoding.EncodedLen(len(sum)))
	base64.StdEncoding.Encode(pin, sum[:])
	return pin, cert.Subject.CommonName, nil
}

// KeyPairWithPin returns PEM encoded Certificate and Key along with an SKPI
// fingerprint of the public key.
// https://blog.afoolishmanifesto.com/posts/golang-self-signed-and-pinned-certs/
func KeyPairWithPin(commonName string) (pemCert []byte, pemKey []byte, pin []byte, err error) {
	bits := 4096
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "rsa.GenerateKey")
	}

	tpl := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: commonName},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(2, 0, 0),
		BasicConstraintsValid: true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}
	derCert, err := x509.CreateCertificate(rand.Reader, &tpl, &tpl, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("x509.CreateCertificate: %w", err)
	}

	buf := &bytes.Buffer{}
	err = pem.Encode(buf, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derCert,
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("pem.Encode: %w", err)
	}

	pemCert = buf.Bytes()

	buf = &bytes.Buffer{}
	err = pem.Encode(buf, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("pem.Encode: %w", err)
	}
	pemKey = buf.Bytes()
	cert, err := x509.ParseCertificate(derCert)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("x509.ParseCertificate: %w", err)
	}

	pubDER, err := x509.MarshalPKIXPublicKey(cert.PublicKey.(*rsa.PublicKey))
	if err != nil {
		return nil, nil, nil, fmt.Errorf("x509.MarshalPKIXPublicKey: %w", err)
	}
	sum := sha256.Sum256(pubDER)
	pin = make([]byte, base64.StdEncoding.EncodedLen(len(sum)))
	base64.StdEncoding.Encode(pin, sum[:])

	return pemCert, pemKey, pin, nil
}
func GenerateCertKey(certPath, keyPath string, commonName string) (err error) {
	pemCert, pemKey, _, err := KeyPairWithPin(commonName)
	if err != nil {
		return err
	}
	if err = os.WriteFile(keyPath, pemKey, 0644); err != nil {
		return err
	}
	if err = os.WriteFile(certPath, pemCert, 0644); err != nil {
		return err
	}
	return nil
}
