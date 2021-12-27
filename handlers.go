package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fiatjaf/bip340"
	"github.com/fiatjaf/relayer"
	. "github.com/stevelacy/daz"
)

var head = H("head",
	H("meta", Attr{"charset": "utf-8"}),
	H("meta", Attr{
		"name":    "viewport",
		"content": "width=device-width, initial-scale=1.0",
	}),
	H("title", "rsslay"),
)

func handleWebpage(w http.ResponseWriter, r *http.Request) {
	items := make([]HTML, 0, 200)
	iter := b.db.NewIter(nil)
	for iter.First(); iter.Valid(); iter.Next() {
		pubkey := string(iter.Key())
		var entity Entity
		if err := json.Unmarshal(iter.Value(), &entity); err != nil {
			continue
		}
		items = append(items, H("tr",
			H("td",
				H("code",
					pubkey),
			),
			H("td",
				H("a", Attr{
					"href": entity.URL}, entity.URL),
			),
		))
	}

	body := H("body",
		H("h1", "rsslay"),
		H("p", "rsslay turns RSS or Atom feeds into ",
			H("a", Attr{
				"href": "https://github.com/fiatjaf/nostr"}, "Nostr"),
			" profiles.",
		),
		H("h2", "How to use"),
		H("ol",
			H("li", "Get the blog URL or RSS or Atom feed URL and paste below;"),
			H("li", "Click the button to get its corresponding public key"),
			H("li", "Add the relay ",
				H("code", "wss://"+b.Domain),
				" to your Nostr client",
			),
			H("li", "Follow the feed's public key from your Nostr client."),
		),
		H("form", Attr{
			"action": "/create",
			"method": "GET",
			"class":  "my-4",
		},
			H("label",
				H("input", Attr{
					"name":        "url",
					"type":        "url",
					"placeholder": "https://.../feed",
				}),
			),
			H("button", "Get Public Key"),
		),

		H("h2", "Some of the existing feeds"),
		H("table", items),
	)

	w.Header().Set("content-type", "text/html")
	w.Write([]byte(
		H("html",
			head,
			body,
		)()))
}

func handleCreateFeed(w http.ResponseWriter, r *http.Request) {
	url := r.URL.Query().Get("url")

	if _, err := parseFeed(url); err != nil {
		w.WriteHeader(400)
		fmt.Fprint(w, "bad feed: "+err.Error())
		return
	}

	sk := privateKeyFromFeed(url)
	s, err := bip340.ParsePrivateKey(sk)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprint(w, "bad private key: "+err.Error())
		return
	}
	pubkey := fmt.Sprintf("%x", bip340.GetPublicKey(s))

	j, _ := json.Marshal(Entity{
		PrivateKey: sk,
		URL:        url,
	})

	if err := b.db.Set([]byte(pubkey), j, nil); err != nil {
		w.WriteHeader(500)
		fmt.Fprint(w, "failure: "+err.Error())
		return
	}

	relayer.Log.Info().Str("url", url).Str("pubkey", pubkey).Msg("saved feed")

	fmt.Fprintf(w, "url   : %s\npubkey: %s", url, pubkey)
	return
}
