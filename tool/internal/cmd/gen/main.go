package main

import (
	"context"
	"flag"

	"github.com/octohelm/gengo/pkg/gengo"
	"github.com/octohelm/x/logr"
	"github.com/octohelm/x/logr/slog"
)

import (
	_ "github.com/octohelm/gengo/devpkg/deepcopygen"
	_ "github.com/octohelm/gengo/devpkg/runtimedocgen"

	_ "github.com/octohelm/courier/devpkg/clientgen"
	_ "github.com/octohelm/courier/devpkg/injectablegen"
	_ "github.com/octohelm/courier/devpkg/operatorgen"
)

func main() {
	flag.Parse()

	c, err := gengo.NewExecutor(&gengo.GeneratorArgs{
		Entrypoint:         flag.Args(),
		OutputFileBaseName: "zz_generated",
		Globals:            map[string][]string{},
	})
	if err != nil {
		panic(err)
	}

	ctx := logr.WithLogger(context.Background(), slog.Logger(slog.Default()))

	if err := c.Execute(ctx, gengo.GetRegisteredGenerators()...); err != nil {
		panic(err)
	}
}
