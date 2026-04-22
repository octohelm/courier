package main

import (
	"context"
	"fmt"
	"os"

	"github.com/octohelm/gengo/pkg/agentskill"
)

func main() {
	if err := (&agentskill.Installer{}).Install(context.Background()); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
		return
	}
}
