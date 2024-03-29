package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/linki/instrumented_http"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
)

func main() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe("127.0.0.1:9099", nil))
	}()

	instrumentedClient := instrumented_http.NewClient(nil, &instrumented_http.Callbacks{
		PathProcessor: instrumented_http.LastPathElementProcessor,
	})

	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithHTTPClient(instrumentedClient))
	if err != nil {
		log.Fatal(err)
	}

	client := route53.NewFromConfig(cfg)

	for {
		zones, err := client.ListHostedZones(context.Background(), &route53.ListHostedZonesInput{})
		if err != nil {
			log.Fatal(err)
		}

		for _, z := range zones.HostedZones {
			fmt.Println(aws.ToString(z.Name))
		}

		time.Sleep(10 * time.Second)
	}
}

// expected result:
//
// $ curl -Ss 127.0.0.1:9099/metrics | grep http
//
// http_request_duration_seconds{handler="instrumented_http",host="route53.amazonaws.com",method="GET",path="hostedzone",query="",scheme="https",status="200",quantile="0.5"} 0.621713504
// http_request_duration_seconds{handler="instrumented_http",host="route53.amazonaws.com",method="GET",path="hostedzone",query="",scheme="https",status="200",quantile="0.9"} 0.871172988
// http_request_duration_seconds{handler="instrumented_http",host="route53.amazonaws.com",method="GET",path="hostedzone",query="",scheme="https",status="200",quantile="0.99"} 0.871172988
// http_request_duration_seconds_sum{handler="instrumented_http",host="route53.amazonaws.com",method="GET",path="hostedzone",query="",scheme="https",status="200"} 2.7609258359999997
// http_request_duration_seconds_count{handler="instrumented_http",host="route53.amazonaws.com",method="GET",path="hostedzone",query="",scheme="https",status="200"} 4
