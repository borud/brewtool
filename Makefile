all: lint vet test build

build: brewtool

brewtool:
	@cd cmd/$@ && go build -o ../../bin/$@

test:
	@go test ./...

vet:
	@go vet ./...

lint:
	@revive ./...

clean:
	@rm -rf bin
