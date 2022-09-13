all: linux-amd64

.ONESHELL:
linux-amd64: fmt
	GOARCH=amd64 GOOS=linux go build  -ldflags '-s -w' -v -o build/wireguard-gui-amd64 .

tool-gofumpt:
	which golangci-lint > /dev/null 2>&1 || go install mvdan.cc/gofumpt@latest

fmt: tool-gofumpt
	gofumpt -w --extra ./
