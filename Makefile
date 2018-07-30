all:
	go build -i -v

test:
	go test -v ./pkg/...

.PHONY: all test
