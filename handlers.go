package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fiatjaf/bip340"
	"github.com/fiatjaf/relayer"
)

func handleWebpage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "hello world\n\n")

	iter := b.db.NewIter(nil)
	for iter.First(); iter.Valid(); iter.Next() {
		pubkey := string(iter.Key())
		var entity Entity
		if err := json.Unmarshal(iter.Value(), &entity); err != nil {
			continue
		}
		fmt.Fprintf(w, "%s %s", pubkey, entity.URL)
	}
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
