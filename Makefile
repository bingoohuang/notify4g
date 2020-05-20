package-standalone:
	./scripts/build.py --package --platform=linux --arch=amd64 --role=standalone

proxy:
	export GOPROXY=https://goproxy.cn

build: proxy
	gofmt -s -w .&&go mod tidy&&go fmt ./...&&revive .&&goimports -w .&&golangci-lint run --enable-all&&go install -ldflags="-s -w" ./...
