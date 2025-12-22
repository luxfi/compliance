.PHONY: build run test test-race lint docker docker-push tag clean

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)

build:
	go build -ldflags="-s -w -X main.version=$(VERSION)" -o complianced ./cmd/complianced/

run: build
	./complianced

test:
	go test -count=1 ./...

test-race:
	go test -race -count=1 ./...

lint:
	go vet ./...

docker:
	docker build --platform linux/amd64 -t ghcr.io/luxfi/compliance:$(VERSION) -t ghcr.io/luxfi/compliance:latest .

docker-push:
	docker push ghcr.io/luxfi/compliance:$(VERSION)
	docker push ghcr.io/luxfi/compliance:latest

tag:
	@echo "Current version: $(VERSION)"
	@echo "Usage: git tag -a v0.X.0 -m 'release v0.X.0' && git push origin v0.X.0"

clean:
	rm -f complianced
