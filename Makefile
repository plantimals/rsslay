rsslay: $(shell find . -name "*.go")
	CC=$$(which musl-gcc) go build -ldflags="-s -w -linkmode external -extldflags '-static'" -o ./rsslay

deploy: rsslay
	ssh root@turgot 'systemctl stop rsslay'
	scp rsslay turgot:rsslay/rsslay
	ssh root@turgot 'systemctl start rsslay'
