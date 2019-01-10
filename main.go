package main

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"os"
	"time"

	"contrib.go.opencensus.io/exporter/stackdriver"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
)

var (
	mRequestLatencyMs = stats.Int64("request_latency", "The latency in milliseconds per request", "ms")
	memoryTypeTag, _  = tag.NewKey("memory_type")
)

func reportLatency(ctx context.Context, start time.Time) {
	elapsed := time.Since(start)
	stats.Record(ctx, mRequestLatencyMs.M(elapsed.Nanoseconds()/1000000))
}

func getDistributionBuckets() []float64 {
	var buckets []float64
	for i := 0; i < 200; i++ {
		buckets = append(buckets, math.Pow(1.1, float64(i)))
	}
	return buckets
}

func enableViews() error {
	requestLatencyView := &view.View{
		Name:        "request_latency",
		Measure:     mRequestLatencyMs,
		Description: "The distribution of request latencies",
		TagKeys:     []tag.Key{memoryTypeTag},
		Aggregation: view.Distribution(getDistributionBuckets()...),
	}
	return view.Register(requestLatencyView)
}

func doWork(ctx context.Context) {
	defer reportLatency(ctx, time.Now())
	sleepTime := rand.ExpFloat64() / .5
	fmt.Printf("Sleeptime: %v", sleepTime)
	time.Sleep(time.Duration(sleepTime) * time.Second)
}

func handler(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	if rand.Intn(2) > 0 {
		tag.Upsert(memoryTypeTag, "high")
	} else {
		tag.Upsert(memoryTypeTag, "low")
	}
	doWork(ctx)
	fmt.Fprintf(w, "Hello, world...")
}

func main() {
	sd, _ := stackdriver.NewExporter(stackdriver.Options{
		ProjectID: "acemil-1",
		Resource: &monitoredrespb.MonitoredResource{
			Type: "gce_instance",
			Labels: map[string]string{
				"instance_id": os.Getenv("MY_GCE_INSTANCE_ID"),
				"zone":        os.Getenv("MY_GCE_INSTANCE_ZONE"),
			},
		},
	})
	defer sd.Flush()
	view.RegisterExporter(sd)
	view.SetReportingPeriod(60 * time.Second)
	enableViews()
	http.HandleFunc("/", handler)
	http.ListenAndServe(":8080", nil)
}
