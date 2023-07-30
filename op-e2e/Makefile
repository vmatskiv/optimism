test:
	go test -v ./...

lint:
	golangci-lint run -E goimports,sqlclosecheck,bodyclose,asciicheck,misspell,errorlint --timeout 5m -e "errors.As" -e "errors.Is" ./...

.PHONY: \
	test \
	lint
