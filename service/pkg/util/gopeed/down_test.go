package gopeed

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDown(t *testing.T) {
	type args struct {
		request *Request
	}
	tests := []struct {
		name string
		args args
	}{
		{
			"proxyee-down.ico",
			args{
				&Request{
					"get",
					"https://raw.githubusercontent.com/proxyee-down-org/proxyee-down/master/front/public/favicon.ico",
					map[string]string{
						"Host":            "github.com",
						"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.121 Safari/537.36",
						"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
						"Referer":         "http://github.com/proxyee-down-org/proxyee-down/releases",
						"Accept-Encoding": "gzip, deflate, br",
						"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
					},
					nil,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requireNetwork(t)
			dest := filepath.Join(t.TempDir(), "favicon.ico")
			err := Down(tt.args.request, dest)
			if err != nil {
				t.Errorf("error down= %v", err)
				return
			}
			// The file is written into dir; reading it from the working
			// directory made this test fail on every machine.
			downMd5 := fileMd5(dest)
			if "8de7a6a2e786861013d61b77b2394012" != downMd5 {
				t.Errorf("error md5= %v", downMd5)
				return
			}
		})
	}
}

func TestResolve(t *testing.T) {
	// os.Setenv("HTTP_PROXY", "http://127.0.0.1:8888")
	type args struct {
		request *Request
	}
	tests := []struct {
		name    string
		args    args
		want    *Response
		wantErr bool
	}{
		{
			"proxyee-down.ico",
			args{
				&Request{
					"get",
					"https://raw.githubusercontent.com/proxyee-down-org/proxyee-down/master/front/public/favicon.ico",
					map[string]string{
						"Host":            "github.com",
						"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; WOW64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/72.0.3626.121 Safari/537.36",
						"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
						"Referer":         "http://github.com/proxyee-down-org/proxyee-down/releases",
						"Accept-Encoding": "gzip, deflate, br",
						"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8",
					},
					nil,
				},
			},
			&Response{"favicon.ico", 116093, true},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requireNetwork(t)
			got, err := Resolve(tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Resolve() = %v, want %v", got, tt.want)
			}
		})
	}
}

func fileMd5(filePath string) string {
	file, _ := os.Open(filePath)

	// Tell the program to call the following function when the current function returns
	defer file.Close()

	// Open a new hash interface to write to
	hash := md5.New()

	// Copy the file in the hash interface and check for any error
	if _, err := io.Copy(hash, file); err != nil {
		return ""
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func TestTemp2(t *testing.T) {
	fmt.Println("123456789"[2:3])
}

// Both tests fetch a file from GitHub and compare it with a fixed checksum, so
// they only mean something with network access and an unchanged upstream file.
// Set V2RAYA_NETWORK_TESTS=1 to run them.
func requireNetwork(t *testing.T) {
	t.Helper()
	if os.Getenv("V2RAYA_NETWORK_TESTS") != "1" {
		t.Skip("set V2RAYA_NETWORK_TESTS=1 to run the tests that download from the network")
	}
}
