package pathpattern

import (
	"testing"

	testingx "github.com/octohelm/x/testing"
)

func TestPathnamePatternWithoutMulti(t *testing.T) {
	p := Parse("/users/{userID}/repos/{repoID}")

	testingx.Expect(t, p, testingx.Equal(
		Segments{lit("users"), named("userID"), lit("repos"), named("repoID")},
	))

	t.Run("should get path values success", func(t *testing.T) {
		params, err := p.PathValues("/users/1/repos/2")

		testingx.Expect(t, err, testingx.Be[error](nil))
		testingx.Expect(t, params["userID"], testingx.Be("1"))
		testingx.Expect(t, params["repoID"], testingx.Be("2"))

		testingx.Expect(t, p.Encode(params), testingx.Be("/users/1/repos/2"))
	})

	t.Run("should not get path values which not matched", func(t *testing.T) {
		_, err := p.PathValues("/not-match")
		testingx.Expect(t, err, testingx.Not(testingx.Be[error](nil)))
	})

	t.Run("should not get path values not full matched", func(t *testing.T) {
		_, err := p.PathValues("/users/1/stars/1")
		testingx.Expect(t, err, testingx.Not(testingx.Be[error](nil)))
	})

	t.Run("should encode with empty missing params", func(t *testing.T) {
		testingx.Expect(t, p.Encode(map[string]string{
			"userID": "1",
		}), testingx.Be("/users/1/repos/-"))
	})
}

func TestPathnamePatternWithMulti(t *testing.T) {
	p := Parse("/v2/{name...}/manifests/{reference}")

	testingx.Expect(t, p, testingx.Equal(
		Segments{lit("v2"), namedMulti("name"), lit("manifests"), named("reference")}),
	)

	t.Run("should get path values success", func(t *testing.T) {
		values, err := p.PathValues("/v2/a/b/c/manifests/v1")

		testingx.Expect(t, err, testingx.Be[error](nil))
		testingx.Expect(t, values["name"], testingx.Be("a/b/c"))
		testingx.Expect(t, values["reference"], testingx.Be("v1"))

		testingx.Expect(t, p.Encode(values), testingx.Be("/v2/a/b/c/manifests/v1"))
	})

	t.Run("should not get path values not full matched", func(t *testing.T) {
		_, err := p.PathValues("/v2/a/b/c/blobs/xxx")
		testingx.Expect(t, err, testingx.Not(testingx.Be[error](nil)))
	})
}

func TestPathnamePatternWithoutParams(t *testing.T) {
	p := Parse("/auth/user")

	testingx.Expect(t, p, testingx.Equal(Segments{lit("auth"), lit("user")}))
	testingx.Expect(t, p.Encode(map[string]string{}), testingx.Be("/auth/user"))
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
