tidy: gen fmt
	go mod tidy

test:
	CGO_ENABLED=0 go test -failfast ./...

test.race:
	CGO_ENABLED=1 go test -v -race ./...

fmt:
	goimports -l -w .

dep:
	go get -u ./...

gen:
	go run ./tool/internal/cmd/gen

serve:
	go run ./example/cmd/example