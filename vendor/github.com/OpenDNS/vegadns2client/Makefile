build:
	go build -v .

test:
	mkdir -p .coverage
	go vet -v
	go get -u github.com/golang/lint/golint
	golint -set_exit_status
	go test -v -cover -coverprofile .coverage/cover.out
	go tool cover -html=.coverage/cover.out -o .coverage/index.html
	@echo ""
	@echo "To see test coverage, try:"
	@echo "open .coverage/index.html"
