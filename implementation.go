package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/fiatjaf/go-nostr/event"
	"github.com/fiatjaf/go-nostr/filter"
	"github.com/fiatjaf/relayer"
	"github.com/kelseyhightower/envconfig"
)

type RSSlay struct {
	Secret string `envconfig:"SECRET" required`
	Domain string `envconfig:"DOMAIN" required`

	updates     chan event.Event
	lastEmitted sync.Map
	db          *pebble.DB
}

func (b *RSSlay) Name() string {
	return "rsslay"
}

func (b *RSSlay) Init() error {
	err := envconfig.Process("", b)
	if err != nil {
		return fmt.Errorf("couldn't process envconfig: %w", err)
	}

	if db, err := pebble.Open("db", nil); err != nil {
		relayer.Log.Fatal().Err(err).Str("path", "db").Msg("failed to open db")
	} else {
		b.db = db
	}

	relayer.Router.Path("/").HandlerFunc(handleWebpage)
	relayer.Router.Path("/create").HandlerFunc(handleCreateFeed)

	go func() {
		time.Sleep(20 * time.Minute)

		filters := relayer.GetListeningFilters()
		relayer.Log.Info().Int("filters active", len(filters)).
			Msg("checking for updates")

		for _, filter := range filters {
			if ((filter.Kind != nil && *filter.Kind == event.KindTextNote) ||
				filter.Kind == nil) &&
				filter.TagProfile == "" && filter.TagEvent == "" && filter.ID == "" {

				for _, pubkey := range filter.Authors {
					if val, closer, err := b.db.Get([]byte(pubkey)); err == nil {
						defer closer.Close()

						var entity Entity
						if err := json.Unmarshal(val, &entity); err != nil {
							relayer.Log.Error().Err(err).Str("key", pubkey).
								Str("val", string(val)).
								Msg("got invalid json from db")
							continue
						}

						feed, err := parseFeed(entity.URL)
						if err != nil {
							relayer.Log.Warn().Err(err).Str("url", entity.URL).
								Msg("failed to parse feed")
							continue
						}

						if filter.Kind == nil || *filter.Kind == event.KindTextNote {
							for _, item := range feed.Items {
								evt := itemToTextNote(pubkey, item)
								last, ok := b.lastEmitted.Load(entity.URL)
								if !ok || last.(uint32) < evt.CreatedAt {
									evt.Sign(entity.PrivateKey)
									b.updates <- evt
									b.lastEmitted.Store(entity.URL, last)
								}
							}
						}
					}
				}
			}
		}

	}()

	return nil
}

func (b *RSSlay) SaveEvent(_ *event.Event) error {
	return errors.New("we don't accept any events")
}

func (b *RSSlay) QueryEvents(filter *filter.EventFilter) ([]event.Event, error) {
	var evts []event.Event

	if filter.ID != "" || filter.TagProfile != "" || filter.TagEvent != "" {
		return evts, nil
	}

	for _, pubkey := range filter.Authors {
		if val, closer, err := b.db.Get([]byte(pubkey)); err == nil {
			defer closer.Close()

			var entity Entity
			if err := json.Unmarshal(val, &entity); err != nil {
				relayer.Log.Error().Err(err).Str("key", pubkey).Str("val", string(val)).
					Msg("got invalid json from db")
				continue
			}

			feed, err := parseFeed(entity.URL)
			if err != nil {
				relayer.Log.Warn().Err(err).Str("url", entity.URL).
					Msg("failed to parse feed")
				continue
			}

			if filter.Kind == nil || *filter.Kind == event.KindSetMetadata {
				evt := feedToSetMetadata(pubkey, feed)

				if filter.Since != 0 && evt.CreatedAt < filter.Since {
					continue
				}
				if filter.Until != 0 && evt.CreatedAt > filter.Until {
					continue
				}

				evt.Sign(entity.PrivateKey)
				evts = append(evts, evt)
			}

			if filter.Kind == nil || *filter.Kind == event.KindTextNote {
				var last uint32 = 0
				for _, item := range feed.Items {
					evt := itemToTextNote(pubkey, item)

					if filter.Since != 0 && evt.CreatedAt < filter.Since {
						continue
					}
					if filter.Until != 0 && evt.CreatedAt > filter.Until {
						continue
					}

					evt.Sign(entity.PrivateKey)

					if evt.CreatedAt > last {
						last = evt.CreatedAt
					}

					evts = append(evts, evt)
				}

				b.lastEmitted.Store(entity.URL, last)
			}
		}
	}

	return evts, nil
}

func (b *RSSlay) InjectEvents() chan event.Event {
	return b.updates
}
