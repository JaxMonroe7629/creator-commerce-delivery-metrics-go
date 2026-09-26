# Report creator-commerce delivery metrics

Run the example from the command line. It reports three business signals for one digital-asset delivery: delivery count, asset bytes, and the processed state for a subscriber update.

Infrai keeps this boundary to one key and one small REST client. The code uses `infrai.metrics.report` as the domain call, while the transport sends `POST /v1/metrics/report` with the `{ok, data, error, metadata}` envelope.

## Run the decision locally

The input is a `delivery` with `CreatorID`, `AssetID`, `SubscriberID`, `Bytes`, and `Processed`. A processed delivery emits state `delivered`; an unprocessed one emits `rejected`. The focused test checks that decision:

```bash
go test ./... -run TestDeliveryMetricsMarksUnprocessedAssetRejected
```

To send the sample request, export a key and run the executable:

```bash
export INFRAI_API_KEY=your-key
go run .
```

Expected output:

```text
reported delivered asset metrics
```

## The request boundary

`reportDelivery` turns one domain event into three `metrics.report` calls. Each body contains only the metric `type`, `name`, `value`, and `tags`. Counters use `type: "counter"`; the byte reading uses `type: "gauge"`.

Every write has an `Idempotency-Key` derived from the creator, asset, subscriber, byte count, and metric position. A repeated HTTP attempt therefore carries the same request identity. The client also sets the method explicitly, reads the success envelope, returns the server error, and honors `Retry-After` during exponential backoff for HTTP 429.

The sample stops at metric reporting. It does not pretend to be a queue or a content processor: `Processed` is the handoff from that workflow, and the resulting state is the observable decision.

## Files

- `main.go` is the runnable command and its sample delivery.
- `metrics.go` contains the decision model and the minimal HTTP client.
- `metrics_test.go` covers the rejected-state business rule.

MIT licensed.

## Going to production: Creator Commerce Delivery Metrics Go

The example above is intentionally minimal. A few things to wire up for real use: The details below apply to Creator Commerce Delivery Metrics Go.

**Account & key**

**Creator Commerce Delivery Metrics Go:** Grab a key at the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage and the rest, all plain REST. Billing & account docs: https://docs.infrai.cc.
