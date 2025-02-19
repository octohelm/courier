module github.com/octohelm/courier

go 1.24.0

tool github.com/octohelm/courier/example/cmd/example

tool (
	github.com/octohelm/courier/tool/internal/cmd/gen
	mvdan.cc/gofumpt
)

require (
	github.com/go-courier/logr v0.3.2
	github.com/octohelm/gengo v0.0.0-20250219103331-fc799ce3110a
	github.com/octohelm/x v0.0.0-20250213100717-a5d72cc790e0
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc
	github.com/go-json-experiment/json v0.0.0-20250213060926-925ba3f173fa
	github.com/juju/ansiterm v1.0.0
	github.com/onsi/gomega v1.36.2
	golang.org/x/net v0.35.0
	golang.org/x/sync v0.11.0
	k8s.io/apimachinery v0.32.1
)

require (
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/lunixbochs/vtclean v1.0.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/mod v0.23.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	golang.org/x/tools v0.30.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	mvdan.cc/gofumpt v0.7.0 // indirect
)
