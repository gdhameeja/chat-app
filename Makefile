build:
	go build ./cmd/chat
runapp: build
	./chat
remote-sync:
	rsync -av ../chat projects:/home/gaurav/projects/golang/personal
count-lines:
	find . -iname \*.go -type f | xargs wc -l
