tidy: gen fmt
	go mod tidy

test:
	CGO_ENABLED=0 go test -failfast ./...

test.race:
	CGO_ENABLED=1 go test -v -race ./...

fmt:
	go tool gofumpt -l -w .

dep:
	go get -u ./...

gen:
	go generate ./...

serve:
	go tool example