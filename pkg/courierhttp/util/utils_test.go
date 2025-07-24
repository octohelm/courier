package util

import (
	"net/http"
	"testing"

	"github.com/octohelm/x/testing/bdd"
)

func TestClientIP(t *testing.T) {
	bdd.FromT(t).When("request with remote addr", func(b bdd.T) {
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "127.0.0.1:80"

		b.Then("get client ip",
			bdd.Equal("127.0.0.1", ClientIP(req)),
		)
	})

	bdd.FromT(t).When("request with header X-Forwarded-For", func(b bdd.T) {
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Forwarded-For", "203.0.113.195, 70.41.3.18, 150.172.238.178")

		b.Then("get got client ip",
			bdd.Equal("203.0.113.195", ClientIP(req)),
		)
	})

	bdd.FromT(t).When("request with header X-Real-IP", func(b bdd.T) {
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-Real-IP", "203.0.113.195")

		b.Then("get client ip",
			bdd.Equal("203.0.113.195", ClientIP(req)),
		)
	})

	bdd.FromT(t).When("request with nothing", func(b bdd.T) {
		req, _ := http.NewRequest(http.MethodGet, "/", nil)

		b.Then("could not get client ip",
			bdd.Equal("", ClientIP(req)),
		)
	})
}
