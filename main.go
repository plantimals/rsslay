package main

import (
	"github.com/fiatjaf/go-nostr/event"
	"github.com/fiatjaf/relayer"
)

var b = &RSSlay{
	updates: make(chan event.Event),
}

func main() {
	relayer.Start(b)
}
