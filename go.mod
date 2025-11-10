module github.com/octohelm/courier

go 1.25.1

tool github.com/octohelm/courier/example/cmd/example

tool (
	github.com/octohelm/courier/tool/internal/cmd/gen
	mvdan.cc/gofumpt
)

require (
	github.com/octohelm/gengo v0.0.0-20250928050614-7aa009184957
	github.com/octohelm/x v0.0.0-20251009020353-8be04f917d90
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc
	github.com/go-json-experiment/json v0.0.0-20250910080747-cc2cfa0554c3
	github.com/juju/ansiterm v1.0.0
	golang.org/x/net v0.46.0
	golang.org/x/sync v0.18.0
	k8s.io/apimachinery v0.34.1
)

require (
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/lunixbochs/vtclean v1.0.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/mod v0.29.0 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	golang.org/x/tools v0.38.0 // indirect
	mvdan.cc/gofumpt v0.9.1 // indirect
)
