module github.com/octohelm/courier

go 1.26.2

tool (
	github.com/octohelm/courier/tool/internal/cmd/fmt
	github.com/octohelm/courier/tool/internal/cmd/gen
	github.com/octohelm/courier/tool/internal/cmd/skills-install
)

require (
	// +skill:gengo-guideline
	github.com/octohelm/gengo v0.0.0-20260609052221-02da451b2cd4
	// +skill:testing-guideline
	github.com/octohelm/x v0.0.0-20260508104609-6b72a870e0d2
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc
	github.com/go-json-experiment/json v0.0.0-20260601182631-00ed12fed2a6
	github.com/juju/ansiterm v1.0.0
	golang.org/x/net v0.55.0
	golang.org/x/sync v0.21.0
	k8s.io/apimachinery v0.36.1
)

require (
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/lunixbochs/vtclean v1.0.0 // indirect
	github.com/mattn/go-colorable v0.1.15 // indirect
	github.com/mattn/go-isatty v0.0.22 // indirect
	golang.org/x/mod v0.37.0 // indirect
	golang.org/x/sys v0.46.0 // indirect
	golang.org/x/text v0.38.0 // indirect
	golang.org/x/tools v0.45.0 // indirect
	mvdan.cc/gofumpt v0.10.0 // indirect
)
