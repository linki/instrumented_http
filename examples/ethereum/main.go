package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/linki/instrumented_http"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/rpc"
)

func main() {
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.ListenAndServe("127.0.0.1:9099", nil))
	}()

	httpClient := instrumented_http.NewClient(nil, &instrumented_http.Callbacks{
		PathProcessor:  instrumented_http.IdentityProcessor,
		QueryProcessor: instrumented_http.IdentityProcessor,
	})

	ethClient, err := rpc.DialHTTPWithClient("https://mainnet.infura.io/v2", httpClient)
	if err != nil {
		log.Fatal(err)
	}
	defer ethClient.Close()

	for {
		var blockHeight hexutil.Uint64
		if err := ethClient.Call(&blockHeight, "eth_blockNumber"); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%d\n", blockHeight)

		time.Sleep(10 * time.Second)
	}
}

// expected result:
//
// $ curl -Ss 127.0.0.1:9099/metrics | grep http
//
// http_request_duration_seconds{handler="instrumented_http",host="mainnet.infura.io",method="POST",path="/v2",query="",scheme="https",status="200",quantile="0.5"} 0.153337228
// http_request_duration_seconds{handler="instrumented_http",host="mainnet.infura.io",method="POST",path="/v2",query="",scheme="https",status="200",quantile="0.9"} 1.07004105
// http_request_duration_seconds{handler="instrumented_http",host="mainnet.infura.io",method="POST",path="/v2",query="",scheme="https",status="200",quantile="0.99"} 1.07004105
// http_request_duration_seconds_sum{handler="instrumented_http",host="mainnet.infura.io",method="POST",path="/v2",query="",scheme="https",status="200"} 2.124165826
// http_request_duration_seconds_count{handler="instrumented_http",host="mainnet.infura.io",method="POST",path="/v2",query="",scheme="https",status="200"} 7
