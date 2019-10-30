all:
	go build -i -v -o vmt

test:
	go test -v ./pkg/...

.PHONY: all test
