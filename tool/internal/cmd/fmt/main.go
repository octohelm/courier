package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/octohelm/gengo/pkg/format"
)

func main() {
	flag.Parse()

	ctx := context.Background()

	p := &format.Project{
		Entrypoint: flag.Args(),
		List:       true,
		Write:      true,
	}
	if err := p.Init(ctx); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
		return
	}

	if err := p.Run(ctx); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
		return
	}
}
