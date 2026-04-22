package pathpattern

import (
	"regexp"
	"testing"

	. "github.com/octohelm/x/testing/v2"
)

func TestNormalizePath(t *testing.T) {
	cases := []struct {
		Path       string
		Normalized string
	}{
		{
			Path:       "/",
			Normalized: "/",
		},
		{
			Path:       "/path/to/x",
			Normalized: "/path/to/x",
		},
		{
			Path:       "/path/with/{param}",
			Normalized: "/path/with/{param}",
		},
		{
			Path:       "/path/with/{wildParam...}",
			Normalized: "/path/with/{wildParam...}",
		},
		{
			Path:       "/path/with/:httprouter-style-param",
			Normalized: "/path/with/{httprouter-style-param}",
		},
		{
			Path:       "/path/with/*httprouter-style-param",
			Normalized: "/path/with/{httprouter-style-param...}",
		},
		{
			Path:       "/path/with/{param}",
			Normalized: "/path/with/{param}",
		},
	}

	for _, c := range cases {
		t.Run(c.Path, func(t *testing.T) {
			Then(t, "路径会被规范化为 courier path pattern 形式",
				Expect(NormalizePath(c.Path), Equal(c.Normalized)),
			)
		})
	}
}

func TestPathnamePatternWithoutMulti(t *testing.T) {
	p := Parse("/users/{userID}/repos/{repoID}")

	Then(t, "普通命名路径段会被正确解析",
		Expect(p, Equal(Segments{lit("users"), named("userID"), lit("repos"), named("repoID")})),
	)

	t.Run("完整匹配时返回路径参数", func(t *testing.T) {
		Then(t, "路径参数可提取且可回写", ExpectMust(func() error {
			params, err := p.PathValues("/users/1/repos/2")
			if err != nil {
				return err
			}
			if params["userID"] != "1" || params["repoID"] != "2" {
				return errPathPattern("unexpected params")
			}
			if p.Encode(params) != "/users/1/repos/2" {
				return errPathPattern("unexpected encoded path")
			}
			return nil
		}))
	})

	t.Run("完全不匹配时返回诊断错误", func(t *testing.T) {
		Then(t, "错误消息会包含实际路径和模式", ExpectDo(func() error {
			_, err := p.PathValues("/not-match")
			return err
		}, ErrorMatch(mustPathPatternRE("^路径不匹配：实际路径 /not-match 不符合模式 /users/\\{userID\\}/repos/\\{repoID\\}$"))))
	})

	t.Run("部分匹配时返回诊断错误", func(t *testing.T) {
		Then(t, "错误消息会指出未完整匹配", ExpectDo(func() error {
			_, err := p.PathValues("/users/1/stars/1")
			return err
		}, ErrorMatch(mustPathPatternRE("^路径不匹配：实际路径 /users/1/stars/1 不符合模式 /users/\\{userID\\}/repos/\\{repoID\\}$"))))
	})

	t.Run("缺失参数时用占位符编码", func(t *testing.T) {
		Then(t, "缺失命名参数会回退为 -",
			Expect(p.Encode(map[string]string{"userID": "1"}), Equal("/users/1/repos/-")),
		)
	})
}

func TestPathnamePatternWithMulti(t *testing.T) {
	p := Parse("/v2/{name...}/manifests/{reference}")

	Then(t, "多段路径参数会被正确解析",
		Expect(p, Equal(Segments{lit("v2"), namedMulti("name"), lit("manifests"), named("reference")})),
	)

	t.Run("多段匹配时返回合并后的参数", func(t *testing.T) {
		Then(t, "多段命名参数可提取且可回写", ExpectMust(func() error {
			values, err := p.PathValues("/v2/a/b/c/manifests/v1")
			if err != nil {
				return err
			}
			if values["name"] != "a/b/c" || values["reference"] != "v1" {
				return errPathPattern("unexpected multi-segment values")
			}
			if p.Encode(values) != "/v2/a/b/c/manifests/v1" {
				return errPathPattern("unexpected multi-segment encode result")
			}
			return nil
		}))
	})

	t.Run("多段路径未完整匹配时返回诊断错误", func(t *testing.T) {
		Then(t, "错误消息会保留多段模式上下文", ExpectDo(func() error {
			_, err := p.PathValues("/v2/a/b/c/blobs/xxx")
			return err
		}, ErrorMatch(mustPathPatternRE("^路径不匹配：实际路径 /v2/a/b/c/blobs/xxx 不符合模式 /v2/\\{name\\.\\.\\.\\}/manifests/\\{reference\\}$"))))
	})
}

func TestPathnamePatternWithoutParams(t *testing.T) {
	p := Parse("/auth/user")

	Then(t, "纯静态路径可以正确解析和编码",
		Expect(p, Equal(Segments{lit("auth"), lit("user")})),
		Expect(p.Encode(map[string]string{}), Equal("/auth/user")),
	)
}

func lit(s string) Segment {
	return segment(s)
}

func named(s string) Segment {
	return namedSegment{name: s}
}

func namedMulti(s string) Segment {
	return namedSegment{name: s, multiple: true}
}

func mustPathPatternRE(pattern string) *regexp.Regexp {
	return regexp.MustCompile(pattern)
}

func errPathPattern(msg string) error {
	return &pathPatternErr{msg: msg}
}

type pathPatternErr struct {
	msg string
}

func (e *pathPatternErr) Error() string {
	return e.msg
}
