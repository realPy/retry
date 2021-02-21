package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	_ "net/http/pprof"
	"time"

	"github.com/realPy/retry"
	retrydb "github.com/realPy/retry/store/fs"
)

var (
	//ErrCronRetryAlways ErrCronRetryAlways error
	ErrCronRetryAlways = errors.New("Cron retry always ;)")
)

var counter int = 0

//MyCustomSuccess MyCustomSuccess
func MyCustomSuccess(i interface{}) error {
	fmt.Printf("Payload Succesfull send %s\n", i.(retry.HTTP).Description)

	return ErrCronRetryAlways

}

//MyCustomFailed MyCustomFailed
func MyCustomFailed(i interface{}, e error) {

	if !errors.Is(e, ErrCronRetryAlways) {
		fmt.Printf("CALLER to %s failed: %s: \n", i.(retry.HTTP).Description, e)
	}

}

//MyCustomFinalFailed MyCustomFinalFailed
func MyCustomFinalFailed(i interface{}, e error) {
	fmt.Printf("Final Call to %s Failed: %s \n", i.(retry.HTTP).Description, e)
}

func main() {

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		io.WriteString(w, "{\"success\":true}")

	}))

	rq := retry.RetryQueue{}

	rq.Init(retrydb.NewRStoreFS("./spool", "slrx_"))

	rq.Register(retry.HTTP{}, MyCustomSuccess, MyCustomFailed, MyCustomFinalFailed)

	rq.Start()

	//collect metric  pprof
	go func() {
		log.Println(http.ListenAndServe("localhost:3000", nil))
	}()

	rq.EnqueueExecAndRetry(retry.HTTP{
		Node:         retry.Node{Name: "call1", Description: "HTTPRequest", MaxRetry: retry.MAXRETRYINFINITE, NoPersist: true},
		URL:          server.URL,
		Method:       "POST",
		HTTPPostData: map[string][]string{"data1": {"yes"}},
		HTTPGetData:  map[string][]string{"no": {"yes"}},
		HTTPHeader:   map[string]string{"User-Agent": "noneman"},
		WaitingTime:  1,
	})

	rq.EnqueueExecAndRetry(retry.DelayedNode(retry.HTTP{
		Node:         retry.Node{Name: "call2", Description: "HTTPRequest", MaxRetry: retry.MAXRETRYINFINITE, NoPersist: true},
		URL:          server.URL,
		Method:       "POST",
		HTTPPostData: map[string][]string{"data2": {"yes"}},
		HTTPGetData:  map[string][]string{"no": {"yes"}},
		HTTPHeader:   map[string]string{"User-Agent": "noneman"},
		WaitingTime:  2,
	}, time.Duration(10)*time.Second))
	//

	time.Sleep(30 * time.Second)
	rq.RemoveByName("call1")
	fmt.Printf("Remove call1\n")

	time.Sleep(2000 * time.Second)

}
