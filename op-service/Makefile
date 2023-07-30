test:
	go test -v ./...

lint:
	golangci-lint run -E goimports,sqlclosecheck,bodyclose,asciicheck,misspell,errorlint --timeout 5m -e "errors.As" -e "errors.Is" ./...

generate-mocks:
	go generate ./...

.PHONY: \
	test \
	lint \
	generate-mocks
