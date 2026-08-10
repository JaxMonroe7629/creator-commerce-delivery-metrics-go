package main

import "testing"

func TestDeliveryMetricsMarksUnprocessedAssetRejected(t *testing.T) {
	metrics := deliveryMetrics(delivery{CreatorID: "c1", AssetID: "a1", SubscriberID: "s1", Bytes: 12})
	if got := metrics[0].Name; got != "creator.asset_delivery.total" {
		t.Fatalf("first metric = %q", got)
	}
	if got := metrics[2].Tags["state"]; got != "rejected" {
		t.Fatalf("state = %q, want rejected", got)
	}
	if metrics[2].Value != 1 {
		t.Fatalf("state value = %d, want 1", metrics[2].Value)
	}
}
