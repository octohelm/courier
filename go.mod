module github.com/octohelm/courier

go 1.26.2

tool (
	github.com/octohelm/courier/tool/internal/cmd/fmt
	github.com/octohelm/courier/tool/internal/cmd/gen
	github.com/octohelm/courier/tool/internal/cmd/skills-install
)

require (
	// +skill:gengo-guideline
	github.com/octohelm/gengo v0.0.0-20260508082215-12ad0e3d1026
	// +skill:testing-guideline
	github.com/octohelm/x v0.0.0-20260508075008-3bb151cb7bd8
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc
	github.com/go-json-experiment/json v0.0.0-20260505212615-e40f80bf6836
	github.com/juju/ansiterm v1.0.0
	golang.org/x/net v0.53.0
	golang.org/x/sync v0.20.0
	k8s.io/apimachinery v0.36.0
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/lunixbochs/vtclean v1.0.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.22 // indirect
	golang.org/x/mod v0.35.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	golang.org/x/tools v0.44.0 // indirect
	mvdan.cc/gofumpt v0.10.0 // indirect
)
