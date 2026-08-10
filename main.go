package main

import (
	"context"
	"fmt"
	"os"
)

func main() {
	client, err := newMetricsClient()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	d := delivery{
		CreatorID: "creator-17", AssetID: "pack-2026-08", SubscriberID: "subscriber-42",
		Bytes: 1843200, Processed: true,
	}
	if err := reportDelivery(context.Background(), client, d); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Println("reported delivered asset metrics")
}
