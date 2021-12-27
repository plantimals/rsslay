module github.com/fiatjaf/rsslay

go 1.16

require (
	github.com/cockroachdb/pebble v0.0.0-20211222161641-06e42cfa82c0
	github.com/fiatjaf/bip340 v1.1.0 // indirect
	github.com/fiatjaf/go-nostr v0.2.1
	github.com/fiatjaf/relayer v1.0.0
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/mmcdole/gofeed v1.1.3
)

replace github.com/fiatjaf/relayer => /home/fiatjaf/comp/relayer