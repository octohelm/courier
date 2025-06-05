module github.com/octohelm/courier

go 1.24.3

tool github.com/octohelm/courier/example/cmd/example

tool (
	github.com/octohelm/courier/tool/internal/cmd/gen
	mvdan.cc/gofumpt
)

require (
	github.com/go-courier/logr v0.3.2
	github.com/octohelm/gengo v0.0.0-20250418031154-d14a228621d4
	github.com/octohelm/x v0.0.0-20250409031213-9c254440c2b8
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc
	github.com/go-json-experiment/json v0.0.0-20250517221953-25912455fbc8
	github.com/juju/ansiterm v1.0.0
	github.com/onsi/gomega v1.37.0
	golang.org/x/net v0.41.0
	golang.org/x/sync v0.15.0
	k8s.io/apimachinery v0.33.1
)

require (
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/lunixbochs/vtclean v1.0.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	golang.org/x/mod v0.25.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	golang.org/x/tools v0.33.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	mvdan.cc/gofumpt v0.8.0 // indirect
)
