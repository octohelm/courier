package main

import (
	"context"
	"os"

	"github.com/go-courier/logr"
	"github.com/go-courier/logr/slog"
	_ "github.com/octohelm/courier/devpkg/clientgen"
	_ "github.com/octohelm/courier/devpkg/injectablegen"
	_ "github.com/octohelm/courier/devpkg/operatorgen"
	_ "github.com/octohelm/gengo/devpkg/deepcopygen"
	_ "github.com/octohelm/gengo/devpkg/runtimedocgen"
	"github.com/octohelm/gengo/pkg/gengo"
)

func main() {
	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	c, err := gengo.NewContext(&gengo.GeneratorArgs{
		Entrypoint: []string{
			cwd,
		},
		OutputFileBaseName: "zz_generated",
		Globals: map[string][]string{
			"gengo:runtimedoc": {},
		},
	})
	if err != nil {
		panic(err)
	}

	ctx := logr.WithLogger(context.Background(), slog.Logger(slog.Default()))

	if err := c.Execute(ctx, gengo.GetRegisteredGenerators()...); err != nil {
		panic(err)
	}
}
