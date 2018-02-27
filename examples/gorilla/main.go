package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/linki/instrumented_http"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	gorilla "github.com/gorilla/http"
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

	_ = instrumentedTransport

	for {
		status, _, resp, err := gorilla.DefaultClient.Get("https://www.google.de/?q=kubernetes", nil)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Close()

		fmt.Printf("%d\n", status.Code)

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
