package v2ray

import "testing"

func TestRunDNSRedirectCommandsReturnsError(t *testing.T) {
	if err := runDNSRedirectCommands("exit 23"); err == nil {
		t.Fatal("DNS 重定向命令失败时应返回错误")
	}
}
