package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	_ "net/http/pprof"
	"time"

	"github.com/realPy/retry"
	retrydb "github.com/realPy/retry/store/fs"
)

func MyCustomSuccess(i interface{}) error {
	fmt.Printf("Payload Succesfull send %s\n", i.(retry.HTTP).Description)
	httpNode := i.(retry.HTTP)

	return fmt.Errorf("success but custom failed retry %d/%d", httpNode.Node.GetAttempt()+1, i.(retry.HTTP).MaxRetry)

}

func MyCustomFailed(i interface{}, e error) {
	fmt.Printf("Call to %s failed: %s: \n", i.(retry.HTTP).Description, e)
}

func MyCustomFinalFailed(i interface{}, e error) {
	fmt.Printf("Final Call to %s Failed: %s \n", i.(retry.HTTP).Description, e)
}

func main() {

	rq := retry.RetryQueue{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		fmt.Printf("Server receive data-->%s\n", body)
		io.WriteString(w, "{\"success\":true}")

	}))

	//collect metric  pprof
	go func() {
		log.Println(http.ListenAndServe("localhost:6666", nil))
	}()

	rq.Init(retrydb.NewRStoreFS("./spool", "slrx_"))
	//enregistrer ici les function de reload pour success
	rq.Register(retry.HTTP{}, MyCustomSuccess, MyCustomFailed, MyCustomFinalFailed)

	rq.Start()

	rq.EnqueueExecAndRetry(retry.HTTP{
		Node:         retry.Node{Description: "HTTPRequest", MaxRetry: 4},
		URL:          server.URL,
		Method:       "POST",
		HTTPPostData: map[string][]string{"data1": {"yes"}},
		HTTPGetData:  map[string][]string{"no": {"yes"}},
		HTTPHeader:   map[string]string{"User-Agent": "noneman"},
		WaitingTime:  10,
	})

	time.Sleep(2000 * time.Second)

}
