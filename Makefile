NAME=max7456

build:
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ${NAME}-windows-64 ./cmd/${NAME}
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ${NAME}-darwin-64 ./cmd/${NAME}
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${NAME}-linux-64 ./cmd/${NAME}

.PHONY: build