module github.com/octohelm/courier

go 1.27.0

tool (
	github.com/octohelm/courier/tool/internal/cmd/fmt
	github.com/octohelm/courier/tool/internal/cmd/gen
	github.com/octohelm/courier/tool/internal/cmd/skills-install
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc
	github.com/juju/ansiterm v1.0.0
	// +skill:gengo-guideline
	github.com/octohelm/gengo v0.0.0-20260821034500-ef88841eb5e4
	// +skill:testing-guideline
	github.com/octohelm/x v0.0.0-20260821032215-38f48df07f8c
	golang.org/x/net v0.58.0
	golang.org/x/sync v0.22.0
	k8s.io/apimachinery v0.36.4
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/lunixbochs/vtclean v1.0.0 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	golang.org/x/mod v0.40.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.41.0 // indirect
	golang.org/x/tools v0.49.0 // indirect
	mvdan.cc/gofumpt v0.11.0 // indirect
)
