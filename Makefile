build:
	go build ./cmd/chat
runapp: build
	./chat
