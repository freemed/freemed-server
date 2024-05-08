VERSION := `date +%FT%T%z`
GOROOT  := /opt/go
BINARY  := freemed

all: clean deps binary

binary:
	@echo "- Building binary version ${VERSION}"
	( cd cmd/freemed-server ; go build -ldflags "-X main.Version=${VERSION}" -v )

deps:
	@echo "- Refreshing dependencies"
	( cd cmd/freemed-server ; go get -v -d ./... )

clean:
	@echo "- Cleaning old build files"
	( cd cmd/freemed-server ; go clean -v )

crosscompile:
	( cd cmd/freemed-server ; \
		GOROOT=${GOROOT} CGO_ENABLED=0 GOOS=linux GOARCH=386 \
		go build -v -ldflags "-X main.Version=${VERSION}" \
			-o ${BINARY}.linux.x86 )
	( cd cmd/freemed-server ; \
		GOROOT=${GOROOT} CGO_ENABLED=0 GOOS=windows GOARCH=386 \
		go build -v -ldflags "-X main.Version=${VERSION}" \
			-o ${BINARY}.x86.exe )
	( cd cmd/freemed-server ; \
		GOROOT=${GOROOT} CGO_ENABLED=0 GOOS=darwin GOARCH=386 \
		go build -v -ldflags "-X main.Version=${VERSION}" \
			-o ${BINARY}.mac.bin )
