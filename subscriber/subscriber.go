package subscriber

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"Occam_project/priceprovider"
	"Occam_project/ticker"
)

const (
	barInterval = time.Minute * 1
)

// Subscription is an interface to get price info from multiple price providers.
type Subscription interface {
	Ticker() ticker.Ticker
	Updates() <-chan ticker.Bar
}

// sub implements Subscription interface
type sub struct {
	ticker  ticker.Ticker
	updates chan ticker.Bar
}

// Updates returns a channel with streamed price updates.
func (s *sub) Updates() <-chan ticker.Bar {
	return s.updates
}

// Ticker returns a ticker on which the subscriber is subscribed.
func (s *sub) Ticker() ticker.Ticker {
	return s.ticker
}

// Subscribe creates a new Subscription which merges all price streams of price providers into a single stream.
func Subscribe(t ticker.Ticker, providers ...priceprovider.PriceStreamSubscriber) Subscription {
	s := sub{
		ticker:  t,
		updates: make(chan ticker.Bar),
	}
	wg := sync.WaitGroup{}
	wg.Add(len(providers))

	for _, p := range providers {
		go func(bar chan ticker.Bar, err chan error) {
			defer wg.Done()
			for b := range bar {
				s.updates <- b
			}
			fmt.Println(<-err) // Just print an error to simplify.
		}(p.SubscribePriceStream(t))
	}

	go func() {
		wg.Wait()
		close(s.updates)
	}()

	return &s
}

// nextInterval returns start time for the current bar and duration to the end of this bar.
func nextInterval() (currentTick time.Time, nextTick time.Time, wait time.Duration) {
	now := time.Now()
	currentTick = now.Truncate(barInterval)
	nextTick = currentTick.Add(barInterval)
	wait = nextTick.Sub(now)
	return
}

// IndexPrice provides a Subscription that collects prices from the source Subscription
// and writes the result in the returned Subscription's output channel as a "bar with predefined interval".
func IndexPrice(subscription Subscription) Subscription {
	s := &sub{
		ticker:  subscription.Ticker(),
		updates: make(chan ticker.Bar),
	}

	go func() {
		values := make(map[time.Time][]*ticker.Bar)
		stream := subscription.Updates()

		// Let's wait until the next interval.
		// Values for the current interval might be waiting in the channel already, but they'll be filtered.
		// We might also get them from the channel and just discard.
		_, _, wait := nextInterval()
		<-time.After(wait)

		currentTick, nextTick, wait := nextInterval()
		t := time.NewTimer(wait)

		timerHandler := func() {
			s.updates <- ticker.Bar{
				Ticker: s.ticker,
				Time:   currentTick,
				Price:  averagePrice(values[currentTick]),
			}
			delete(values, currentTick)

			currentTick, nextTick, wait = nextInterval()
			t.Reset(wait)
		}

		for {
			select {
			// If the timer fires, we will handle it after at most one iteration which is ok in our case.
			case <-t.C:
				timerHandler()
			default:
				select {
				case <-t.C:
					timerHandler()
				case value, ok := <-stream:
					if !ok {
						close(s.updates)
						if !t.Stop() {
							<-t.C
						}
						return
					}
					// If we get outdated values, we just ignore them because we try to provide data in real time.
					// If we get data for the next interval (price came in channel earlier than timer fired),
					// we preserve them, but do not consider in the current interval, though in the next one we do.
					if !value.Time.Before(currentTick) {
						if value.Time.Before(nextTick) {
							values[currentTick] = append(values[currentTick], &value)
						} else {
							if value.Time.Before(nextTick.Add(barInterval)) {
								values[nextTick] = append(values[nextTick], &value)
							} else {
								// I have no idea how did we get here.
							}
						}
					}
				}
			}
		}
	}()

	return s
}

// averagePrice returns an average value of the price for all Bars in the list.
func averagePrice(values []*ticker.Bar) string {
	var total float64
	length := len(values)

	if length == 0 {
		return ""
	}

	for _, i := range values {
		price, _ := strconv.ParseFloat(i.Price, 64)
		total += price
	}

	return fmt.Sprintf("%g", total/float64(length))
}
