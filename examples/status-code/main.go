package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/linki/instrumented_http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe("127.0.0.1:9099", nil))
	}()

	client := instrumented_http.NewClient(nil, &instrumented_http.Callbacks{
		PathProcessor:  instrumented_http.IdentityProcessor,
		QueryProcessor: instrumented_http.IdentityProcessor,
		CodeProcessor: func(code int) string {
			// export all status codes >= 400 as label status=failure instead of their individual values.
			if code >= http.StatusBadRequest {
				return "failure"
			}
			return "success"
		},
	})

	for {
		func() {
			resp, err := client.Get("https://kubernetes.io/docs/search/?q=pods")
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()

			fmt.Printf("%d\n", resp.StatusCode)
		}()

		time.Sleep(10 * time.Second)
	}
}

// expected result:
//
// $ curl -Ss 127.0.0.1:9099/metrics | grep http
//
// http_request_duration_seconds{handler="instrumented_http",host="kubernetes.io",method="GET",path="/docs/search/",query="q=pods",scheme="https",status="success",quantile="0.5"} 0.662351
// http_request_duration_seconds{handler="instrumented_http",host="kubernetes.io",method="GET",path="/docs/search/",query="q=pods",scheme="https",status="success",quantile="0.9"} 1.000437
// http_request_duration_seconds{handler="instrumented_http",host="kubernetes.io",method="GET",path="/docs/search/",query="q=pods",scheme="https",status="success",quantile="0.99"} 1.000437
// http_request_duration_seconds_sum{handler="instrumented_http",host="kubernetes.io",method="GET",path="/docs/search/",query="q=pods",scheme="https",status="success"} 1.662788442
// http_request_duration_seconds_count{handler="instrumented_http",host="kubernetes.io",method="GET",path="/docs/search/",query="q=pods",scheme="https",status="success"} 2
