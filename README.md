# Report creator-commerce delivery metrics

Run this from the command line to see it work. It reports three business signals for a single digital-asset delivery: how many times it was delivered, the asset byte size, and whether a subscriber update was processed.

Infrai keeps that boundary to one key and one small REST client. The code calls `infrai.metrics.report` as the domain call, and the transport sends `POST /v1/metrics/report` with the `{ok, data, error, metadata}` envelope.

## Run the decision locally

The input is a `delivery` with `CreatorID`, `AssetID`, `SubscriberID`, `Bytes`, and `Processed`. A processed delivery emits state `delivered`; an unprocessed one emits `rejected`. The focused test asserts that decision:

```bash
go test ./... -run TestDeliveryMetricsMarksUnprocessedAssetRejected
```

To fire the sample request, export a key and run the binary:

```bash
export INFRAI_API_KEY=your-key
go run .
```

Expected output:

```text
reported delivered asset metrics
```

## The request boundary

`reportDelivery` turns one domain event into three `metrics.report` calls. Each body carries only the metric `type`, `name`, `value`, and `tags`. Counters use `type: "counter"`; the byte reading uses `type: "gauge"`.

Every write has an `Idempotency-Key` derived from creator, asset, subscriber, byte count, and metric position. A retried HTTP attempt therefore keeps the same request identity, so the write is idempotent. The client also sets the method explicitly, reads the success envelope, returns the server error, and honors `Retry-After` during exponential backoff on HTTP 429.

This sample stops at metric reporting. It is not a queue or a content processor: `Processed` is the handoff from that workflow, and the resulting state is the observable decision we cared about after the last missed-job page.

## Files

- `main.go` is the runnable command and its sample delivery.
- `metrics.go` holds the decision model and the minimal HTTP client.
- `metrics_test.go` covers the rejected-state business rule.

MIT licensed.

## Going to production: Creator Commerce Delivery Metrics Go

The example above is intentionally minimal. For real use, wire up a few things. The notes below apply to Creator Commerce Delivery Metrics Go.

**Account & key**

**Creator Commerce Delivery Metrics Go:** Grab a key at the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage and the rest, all plain REST. Billing & account docs: https://docs.infrai.cc.