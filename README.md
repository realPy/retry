# Retry Go

Retry is a go lib  which helps to implement retrying execution mechanics with persistence.  


In many situations not everything goes as planned. Retry is designed to run a feature and retry as many times as necessary if it fails.  
To retry is also to try each time. Retry is designed to retry indefinitely and implement scheduled job (cron) mechanisms, such as hourly automatic data purging etc.  


You can easily implement HTTP POST / GET sends, execute a command, create an email sending mechanic and never miss an event!  

As a program must be updated, a persistence in Filesystems, badgerDB, gorm is available to retry exactly where it stopped.  

## How to use with http call
```shellscript
go get github.com/realPy/retry
```
import the lib with:  
```shellscript
import "github.com/realPy/retry"
```
Import the storage manager need:  
```shellscript
import retrydb "github.com/realPy/retry/store/fs"
```


Create the queue  
```shellscript
rq := retry.RetryQueue{}
```
Init you queue wity a storage manager  
```shellscript
rq.Init(retrydb.NewRStoreFS("./spool", "slrx_"))
```
Start the queue  
```shellscript
rq.Start()
```
Enqueue the data to POST with and print Success exec if the exec sucessfull
```shellscript
rq.EnqueueExecAndRetry(retry.HTTP{
    Node:         retry.Node{Description: "HTTPRequest", MaxRetry: 4},
    URL:          server.URL,
    Method:       "POST",
    HTTPPostData: map[string][]string{"data1": {"yes"}},
    HTTPGetData:  map[string][]string{"no": {"yes"}},
    HTTPHeader:   map[string]string{"User-Agent": "noneman"},
    WaitingTime:  10,
    SuccessFunc: func(h *retry.HTTP) error {
			fmt.Printf("Success exec\n")
			return nil
	},
})
```

Delayed POST 
```shellscript
	rq.EnqueueExecAndRetry(retry.DelayedNode(retry.HTTP{
		Node:         retry.Node{Description: "HTTPRequest", MaxRetry: 4},
		URL:          server.URL,
		Method:       "POST",
		HTTPPostData: map[string][]string{"data1": {"yes"}},
		HTTPGetData:  map[string][]string{"no": {"yes"}},
		HTTPHeader:   map[string]string{"User-Agent": "noneman"},
		WaitingTime:  10,
		SuccessFunc: func(h *retry.HTTP) error {
			fmt.Printf("Success exec\n")
			return nil
		}}, time.Duration(10)*time.Second),
	)
```


## How to implement my own retry
Implement your own retry is easy.  
You need to use struct embbed Node that contain minimum information the system need  
You can extended your struct with the elements of your need  

Function implementations:  
The lib use persist to save your jobs on disk or database. The lib known how the Node is struct but not how yout struct is.  
You Must implement the interface BinaryMarshaler ( MarshalBinary() (data []byte, err error) ) and BinaryUnMarshaler (UnmarshalBinary(data []byte) error)  
Dont forget to call the encode et decode function of Node and your own encode and decode follow you struct.  
(see http implementation)  
Implement the nodeRetry interface and voila :)  
```shellscript
type NodeRetry interface {
	Init()
	Execute(ctx context.Context) error
	GetData() Node
	SetData(Node) NodeRetry
	OnSuccess() error
	OnFailed(error) error
	TimeToSleepOnNextFailed(int) time.Duration
	OnFinalFailed(error) error
}
```
## How to implement a scheduled job
Scheduled job is a retry infinite that always failed :)  
To implement a "cron like" job, pass the parameter MaxRetry to retry.MAXRETRYINFINITE and  NoPersist to true.  
Dont use persist on sheduled job or you will stack of them on each restart of your service.  
OnSuccess method interface always return an error (ErrCronRetryAlways is for that) even the execute function success. Return an error on success ensure to retry your event  
