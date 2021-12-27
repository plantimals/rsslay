package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/fiatjaf/go-nostr/event"
	"github.com/mmcdole/gofeed"
)

var fp = gofeed.NewParser()

type Entity struct {
	PrivateKey string
	URL        string
}

func parseFeed(url string) (*gofeed.Feed, error) {
	feed, err := fp.ParseURL(url)
	return feed, err
}

func feedToSetMetadata(pubkey string, feed *gofeed.Feed) event.Event {
	metadata := map[string]string{
		"name":  feed.Title,
		"about": feed.Description + "\n\n" + feed.Link,
	}
	if feed.Image != nil {
		metadata["picture"] = feed.Image.URL
	}
	content, _ := json.Marshal(metadata)

	createdAt := time.Now()
	if feed.PublishedParsed != nil {
		createdAt = *feed.PublishedParsed
	}

	evt := event.Event{
		PubKey:    pubkey,
		CreatedAt: uint32(createdAt.Unix()),
		Kind:      event.KindSetMetadata,
		Tags:      event.Tags{},
		Content:   string(content),
	}
	evt.ID = string(evt.Serialize())

	return evt
}

func itemToTextNote(pubkey string, item *gofeed.Item) event.Event {
	content := ""
	if item.Title != "" {
		content = "**" + item.Title + "**\n\n"
	}

	if item.Description != "" {
		content += item.Description
	} else {
		content += item.Description
	}

	if len(content) > 200 {
		content += content[0:199] + "…"
	}
	content += "\n\n" + item.Link

	createdAt := time.Now()
	if item.UpdatedParsed != nil {
		createdAt = *item.UpdatedParsed
	}
	if item.PublishedParsed != nil {
		createdAt = *item.PublishedParsed
	}

	evt := event.Event{
		PubKey:    pubkey,
		CreatedAt: uint32(createdAt.Unix()),
		Kind:      event.KindTextNote,
		Tags:      event.Tags{},
		Content:   content,
	}
	evt.ID = string(evt.Serialize())

	return evt
}

func privateKeyFromFeed(url string) string {
	m := hmac.New(sha256.New, []byte(b.Secret))
	m.Write([]byte(url))
	r := m.Sum(nil)
	return hex.EncodeToString(r)
}
