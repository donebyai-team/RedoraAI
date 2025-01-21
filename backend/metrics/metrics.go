package metrics

import "github.com/streamingfast/dmetrics"

var MetricSet = dmetrics.NewSet()

var GoogleAPIRequestCount = MetricSet.NewCounter("google_api_client_request_count", "Number of requests made to the google API")
