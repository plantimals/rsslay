package main

import (
	"github.com/fiatjaf/go-nostr"
	"github.com/fiatjaf/relayer"
)

var b = &RSSlay{
	updates: make(chan nostr.Event),
}

func main() {
	relayer.Start(b)
}
