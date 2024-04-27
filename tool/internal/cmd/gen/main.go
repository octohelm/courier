package main

import (
	"context"

	"github.com/octohelm/gengo/pkg/gengo"

	_ "github.com/octohelm/courier/devpkg/clientgen"
	_ "github.com/octohelm/courier/devpkg/operatorgen"
	_ "github.com/octohelm/gengo/devpkg/deepcopygen"
	_ "github.com/octohelm/gengo/devpkg/runtimedocgen"
	_ "github.com/octohelm/storage/devpkg/enumgen"
)

func main() {
	c, err := gengo.NewContext(&gengo.GeneratorArgs{
		Entrypoint: []string{
			"github.com/octohelm/courier/example/client/example",
			"github.com/octohelm/courier/example/apis",
		},
		OutputFileBaseName: "zz_generated",
		Globals: map[string][]string{
			"gengo:runtimedoc": {},
		},
	})
	if err != nil {
		panic(err)
	}

	if err := c.Execute(context.Background(), gengo.GetRegisteredGenerators()...); err != nil {
		panic(err)
	}
}
