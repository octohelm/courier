package testingutil

import (
	"testing"

	testingx "github.com/octohelm/x/testing"
)

func Expect[A any](t testing.TB, actual A, matcheres ...testingx.Matcher[A]) {
	t.Helper()
	testingx.Expect[A](t, actual, matcheres...)
}

func Be[A any](e A) testingx.Matcher[A] {
	return testingx.Be(e)
}

func NotBe[A any](e A) testingx.Matcher[A] {
	return testingx.Not(testingx.Be(e))
}

func Equal[A any](e A) testingx.Matcher[A] {
	return testingx.Equal(e)
}

func NotEqual[A any](e A) testingx.Matcher[A] {
	return testingx.Not(testingx.Equal(e))
}

func HaveLen[A any](c int) testingx.Matcher[A] {
	return testingx.HaveLen[A](c)
}

func NotHaveLen[A any](c int) testingx.Matcher[A] {
	return testingx.Not(testingx.HaveLen[A](c))
}
