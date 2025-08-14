module github.com/octohelm/courier

go 1.25.0

tool github.com/octohelm/courier/example/cmd/example

tool (
	github.com/octohelm/courier/tool/internal/cmd/gen
	mvdan.cc/gofumpt
)

require (
	github.com/go-courier/logr v0.3.2
	github.com/octohelm/gengo v0.0.0-20250711045910-061ca3315825
	github.com/octohelm/x v0.0.0-20250814014249-e79c91160fc4
)

require (
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc
	github.com/go-json-experiment/json v0.0.0-20250813233538-9b1f9ea2e11b
	github.com/juju/ansiterm v1.0.0
	golang.org/x/net v0.43.0
	golang.org/x/sync v0.16.0
	k8s.io/apimachinery v0.33.4
)

require (
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/lunixbochs/vtclean v1.0.0 // indirect
	github.com/mattn/go-colorable v0.1.14 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/onsi/gomega v1.38.0 // indirect
	golang.org/x/mod v0.27.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	golang.org/x/tools v0.36.0 // indirect
	mvdan.cc/gofumpt v0.8.0 // indirect
)
