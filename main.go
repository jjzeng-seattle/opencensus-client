package main

import (
	"context"
	// "fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"contrib.go.opencensus.io/exporter/ocagent"
	"go.opencensus.io/resource"
	"go.opencensus.io/stats"
	"go.opencensus.io/stats/view"
	"go.opencensus.io/tag"
	"go.opencensus.io/zpages"
)

func main() {
	// For the purposes of debugging, we'll add zPages that you can
	// use as a diagnostic to examine if stats is exported
	// out. You can learn about using zPages at https://opencensus.io/zpages/go/
	zPagesMux := http.NewServeMux()
	zpages.Handle(zPagesMux, "/debug")
	go func() {
		if err := http.ListenAndServe(":9999", zPagesMux); err != nil {
			log.Fatalf("Failed to serve zPages")
		}
	}()

	opts := []ocagent.ExporterOption{ocagent.WithServiceName("revision")}
	opts = append(opts, ocagent.WithAddress("localhost:55678"))
	opts = append(opts, ocagent.WithReconnectionPeriod(5*time.Second))
	opts = append(opts, ocagent.WithInsecure())
	opts = append(opts, ocagent.WithResourceDetector(func(context.Context) (*resource.Resource, error) {
		return &resource.Resource{
			Type: "knative_revision",
			Labels: map[string]string{
				"project_id": "jjzeng-knative-dev",
				// "location":           "us-central1-a",
				"cloud.zone": "us-central1-a",
				// "cluster_name":       "purple",
				"k8s.cluster.name":   "purple",
				"service_name":       "helloworld-go",
				"revision_name":      "helloworld-go-hfc7j",
				"configuration_name": "helloworld-go",
				"namespace_name":     "default",
			},
		}, nil
	}))
	oce, err := ocagent.NewExporter(opts...)
	if err != nil {
		log.Fatalf("Failed to create ocagent-exporter: %v", err)
	}

	view.RegisterExporter(oce)

	// Some configurations to get observability signals out.
	view.SetReportingPeriod(60 * time.Second)

	// Some stats
	mPodCounts := stats.Int64("actual_pods", "The total number of knative pods", stats.UnitDimensionless)
	mRequestCount := stats.Int64("request_count", "The total number of requests to revistions", stats.UnitDimensionless)

	myKey := tag.MustNewKey("foo")
	respnseCodeClassKey := tag.MustNewKey("response_code_class")
	respnseCodeKey := tag.MustNewKey("response_code")
	views := []*view.View{
		{
			Description: "The total number of knative pods",
			Name:        "autoscaler/actual_pods",
			Measure:     mPodCounts,
			Aggregation: view.LastValue(),
			TagKeys:     []tag.Key{myKey},
		},
		{
			Description: "The total number of requests to revisions",
			Name:        "revision/request_count",
			Measure:     mRequestCount,
			Aggregation: view.Sum(),
			TagKeys:     []tag.Key{myKey, respnseCodeClassKey, respnseCodeKey},
		},
	}

	if err := view.Register(views...); err != nil {
		log.Fatalf("Failed to register views for metrics: %v", err)
	}

	podCountsCtx := context.Background()
	podCountsCtx, _ = tag.New(podCountsCtx, tag.Insert(myKey, "bar"))
	responseCodeCtx := context.Background()
	responseCodeCtx, _ = tag.New(responseCodeCtx, tag.Insert(respnseCodeClassKey, "4xx"))
	responseCodeCtx, _ = tag.New(responseCodeCtx, tag.Insert(respnseCodeKey, "400"))
	responseCodeCtx, _ = tag.New(responseCodeCtx, tag.Insert(myKey, "bar"))
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	for {
		time.Sleep(5000 * time.Millisecond)
		randPodCounts := rng.Int63n(999)
		randRequestCount := rng.Int63n(100)
		stats.Record(podCountsCtx, mPodCounts.M(randPodCounts))
		stats.Record(responseCodeCtx, mRequestCount.M(randRequestCount))
	}
}
