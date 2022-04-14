package priceprovider

import (
	"Occam_project/ticker"
	"fmt"
	"github.com/pkg/errors"
	"math/rand"
	"time"
)

// PriceStreamSubscriber is an interface to subscribe on a price provider's channel.
type PriceStreamSubscriber interface {
	SubscribePriceStream(ticker ticker.Ticker) (chan ticker.Bar, chan error)
}

// PriceProvider is a mock implementation of PriceStreamSubscriber.
type PriceProvider struct {
	name  string
	price chan ticker.Bar
	err   chan error
}

// New creates a brand-new PriceProvider.
func New(name string) *PriceProvider {
	return &PriceProvider{
		name:  name,
		price: make(chan ticker.Bar),
		err:   make(chan error, 1),
	}
}

// SubscribePriceStream returns two channels: one for prices and one for errors.
// Starts a goroutine to simulate price stream from an exchange.
func (pp *PriceProvider) SubscribePriceStream(tickerName ticker.Ticker) (chan ticker.Bar, chan error) {
	go func() {
		rand.Seed(time.Now().UnixNano())
		tick := time.NewTicker(time.Second * 1) // Send in the price channel periodically.
		done := time.After(time.Second * 180)   // Simulate an error after some time.
		for {
			select {
			case <-tick.C:
				pp.price <- ticker.Bar{
					Ticker: tickerName,
					Time:   time.Now().UTC(),
					Price:  fmt.Sprintf("%g", 98+rand.Float64()*4),
				}
			case <-done:
				tick.Stop()
				pp.err <- errors.Errorf("an error occurred in the stream %s", pp.name)
				close(pp.price)
				return
			}
		}
	}()
	return pp.price, pp.err
}
