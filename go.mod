module github.com/octohelm/courier

go 1.25.1

tool github.com/octohelm/courier/example/cmd/example

tool (
	github.com/octohelm/courier/tool/internal/cmd/gen
	mvdan.cc/gofumpt
)

require (
	github.com/go-courier/logr v0.3.2
	github.com/octohelm/gengo v0.0.0-20250909020815-1e94629296bc
	github.com/octohelm/x v0.0.0-20250905103750-d1a271ae07dd
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc
	github.com/go-json-experiment/json v0.0.0-20250910080747-cc2cfa0554c3
	github.com/juju/ansiterm v1.0.0
	golang.org/x/net v0.44.0
	golang.org/x/sync v0.17.0
	k8s.io/apimachinery v0.34.1
)

require (
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/lunixbochs/vtclean v1.0.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/mod v0.28.0 // indirect
	golang.org/x/sys v0.36.0 // indirect
	golang.org/x/text v0.29.0 // indirect
	golang.org/x/tools v0.37.0 // indirect
	mvdan.cc/gofumpt v0.9.1 // indirect
)
