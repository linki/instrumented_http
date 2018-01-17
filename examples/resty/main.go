package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/linki/instrumented_http"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/go-resty/resty"
)

func main() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe("127.0.0.1:9099", nil))
	}()

	instrumentedTransport := instrumented_http.NewTransport(nil, &instrumented_http.Callbacks{
		PathProcessor:  instrumented_http.IdentityProcessor,
		QueryProcessor: instrumented_http.IdentityProcessor,
	})

	client := resty.New().SetTransport(instrumentedTransport)

	for {
		resp, err := client.R().Get("https://kubernetes.io/docs/search/?q=pods")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%d\n", resp.StatusCode())

		time.Sleep(10 * time.Second)
	}
}

// expected result:
//
// $ curl -Ss 127.0.0.1:9099/metrics | grep http
//
// http_request_duration_seconds{handler="instrumented_http",host="kubernetes.io",method="GET",path="/docs/search/",query="q=pods",scheme="https",status="200",quantile="0.5"} 0.662351
// http_request_duration_seconds{handler="instrumented_http",host="kubernetes.io",method="GET",path="/docs/search/",query="q=pods",scheme="https",status="200",quantile="0.9"} 1.000437
// http_request_duration_seconds{handler="instrumented_http",host="kubernetes.io",method="GET",path="/docs/search/",query="q=pods",scheme="https",status="200",quantile="0.99"} 1.000437
// http_request_duration_seconds_sum{handler="instrumented_http",host="kubernetes.io",method="GET",path="/docs/search/",query="q=pods",scheme="https",status="200"} 1.662788415
// http_request_duration_seconds_count{handler="instrumented_http",host="kubernetes.io",method="GET",path="/docs/search/",query="q=pods",scheme="https",status="200"} 2