package main

import (
	"fmt"
	"time"

	"github.com/realPy/retry"
	retrydb "github.com/realPy/retry/store/fs"
)

func MyCustomSuccess(i interface{}) error {
	fmt.Printf("succes execute\n")

	//return fmt.Errorf("success but custom failed retry %d/%d", i.(retryit.Node).GetAttempt()+1, i.(retryit.Node).MaxRetry)
	return fmt.Errorf("But want retry anyway")
}

func MyCustomFailed(i interface{}, e error) {
	fmt.Printf("Call to %s failed: %s: \n", i.(retry.Node).Description, e)
}

func MyCustomFinalFailed(i interface{}, e error) {
	fmt.Printf("Final Call to %s Failed: %s \n", i.(retry.Node).Description, e)
}

func main() {

	rq := retry.RetryQueue{}

	rq.Init(retrydb.NewRStoreFS("./spool", "slrx_"))
	//register here global functions
	rq.Register(retry.Node{}, MyCustomSuccess, MyCustomFailed, MyCustomFinalFailed)

	rq.Start()

	rq.EnqueueExecAndRetry(retry.Node{Description: "Execute test", MaxRetry: 4})

	time.Sleep(2000 * time.Second)

}
