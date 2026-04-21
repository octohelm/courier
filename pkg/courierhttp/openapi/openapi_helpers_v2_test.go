package openapi

import (
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

func TestNamingAndPkgNamingPrefix(t0 *testing.T) {
	Then(t0, "命名辅助方法符合预期",
		ExpectMust(func() error {
			opt := &buildOption{}
			Naming(func(s string) string { return "custom:" + s })(opt)
			if opt.naming == nil || opt.naming("demo") != "custom:demo" {
				return errOpenAPI("unexpected naming option")
			}
			return nil
		}),
		ExpectMust(func() error {
			prefixes := PkgNamingPrefix{}
			prefixes.Register("github.com/example/project/pkg", "demo")
			if got := prefixes.Prefix("github.com/example/project/pkg/user", "profile"); got != "DemoProfile" {
				return errOpenAPI("unexpected prefix result: " + got)
			}
			if got := prefixes.Prefix("github.com/other/project", "profile"); got != "Profile" {
				return errOpenAPI("unexpected default prefix result: " + got)
			}
			return nil
		}),
		ExpectMust(func() error {
			RegisterPkgNamingPrefix("github.com/example/courierhttp/openapi/helpers", "helper")
			if got := defaultPkgNamingPrefix.Prefix("github.com/example/courierhttp/openapi/helpers/sub", "name"); got != "HelperName" {
				return errOpenAPI("unexpected registered prefix result: " + got)
			}
			return nil
		}),
	)
}

func TestRuntimeDocHelpers(t0 *testing.T) {
	Then(t0, "运行时文档辅助方法可用",
		ExpectMust(func() error {
			if doc, ok := new(BuildFunc).RuntimeDoc(); !ok || len(doc) != 0 {
				return errOpenAPI("unexpected BuildFunc doc")
			}
			if doc, ok := new(BuildOptionFunc).RuntimeDoc(); !ok || len(doc) != 0 {
				return errOpenAPI("unexpected BuildOptionFunc doc")
			}
			if doc, ok := new(PkgNamingPrefix).RuntimeDoc(); !ok || len(doc) != 0 {
				return errOpenAPI("unexpected PkgNamingPrefix doc")
			}
			if _, ok := runtimeDoc(struct{}{}, "", "name"); ok {
				return errOpenAPI("unexpected runtimeDoc hit")
			}
			return nil
		}),
	)
}

func errOpenAPI(msg string) error {
	return &openapiHelperError{msg: msg}
}

type openapiHelperError struct {
	msg string
}

func (e *openapiHelperError) Error() string {
	return e.msg
}
