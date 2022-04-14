package main

import (
	"fmt"

	"Occam_project/priceprovider"
	"Occam_project/subscriber"
	"Occam_project/ticker"
)

func main() {
	// Create some price providers.
	var providers []priceprovider.PriceStreamSubscriber
	for i := 0; i < 100; i++ {
		providers = append(providers, priceprovider.New(fmt.Sprintf("provider_%d", i)))
	}

	// Get Subscription and wrap it with IndexPrice to get a fair price.
	subscription := subscriber.IndexPrice(subscriber.Subscribe(ticker.BTCUSDTicker, providers...))

	fmt.Println("Timestamp", "IndexPrice")

	// Now let's get our fair price from the channel.
	for tickerPrice := range subscription.Updates() {
		fmt.Println(tickerPrice.Time.Unix(), tickerPrice.Price)
	}

}
