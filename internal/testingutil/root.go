package testingutil

import (
	"os"
	"path/filepath"
)

func ProjectRoot() string {
	p, _ := os.Getwd()
	for {
		if p == "/" {
			break
		}
		if _, err := os.Stat(filepath.Join(p, "go.mod")); err == nil {
			return p
		}
		p = filepath.Dir(p)
	}
	return p
}
