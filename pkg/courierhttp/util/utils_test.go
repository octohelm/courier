package util

import (
	"net/http"
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

func TestClientIP(t *testing.T) {
	t.Run("falls back to remote addr", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "127.0.0.1:80"
		Then(t, "ClientIP 返回 remote addr 中的主机地址", Expect(ClientIP(req), Equal("127.0.0.1")))
	})

	t.Run("prefers X-Forwarded-For", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Forwarded-For", "203.0.113.195, 70.41.3.18, 150.172.238.178")
		Then(t, "ClientIP 优先返回第一个 X-Forwarded-For 地址", Expect(ClientIP(req), Equal("203.0.113.195")))
	})

	t.Run("uses X-Real-IP when forwarded header absent", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Real-IP", "203.0.113.195")
		Then(t, "ClientIP 返回 X-Real-IP", Expect(ClientIP(req), Equal("203.0.113.195")))
	})

	t.Run("returns empty when no source exists", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		Then(t, "缺少来源时返回空字符串", Expect(ClientIP(req), Equal("")))
	})
}
