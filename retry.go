package retry

import (
	"bytes"
	"container/list"
	"encoding/gob"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"time"

	"github.com/rs/xid"
)

var (
	//MAXRETRYINFINITE generate infinite retry (cron)
	MAXRETRYINFINITE = -1
)

var (
	//ErrRetryNotImplemented return error when default behaviour
	ErrRetryItNotImplemented = errors.New("The method is not implemented")
)

//RStore RStore struct
type RStore interface {
	ParseAll(func(string, []byte) error)
	Delete(string) error
	Store(string, func() (*bytes.Buffer, error)) error
}

//NodeRetry NodeRetry struct
type NodeRetry interface {
	Init()
	Execute() error
	GetData() Node
	SetData(Node) NodeRetry
	OnSuccess() error
	OnFailed(error) error
	TimeToSleepOnNextFailed(int) time.Duration
	OnFinalFailed(error) error
}

//RetryQueue RetryQueue struct
type RetryQueue struct {
	//waitingList             []NodeRetry
	waitingNodeList         *list.List
	registerSuccessFunc     map[string]func(interface{}) error
	registerFailedFunc      map[string]func(interface{}, error)
	registerFinalFailedFunc map[string]func(interface{}, error)
	io                      chan interface{}
	stop                    chan bool
	wg                      *sync.WaitGroup
	store                   RStore
	nextTriggerTime         time.Time
	latency                 int
}

//delayedNode delayedNode struct
type delayedNode struct {
	node interface{}
	time time.Duration
}

//DelayedNode Create a Delayed node container
func DelayedNode(node interface{}, time time.Duration) interface{} {
	return delayedNode{node: node, time: time}
}

//NodeRemove type of struct to remove node by name
type NodeRemove struct {
	name string
}

//Node Node struct
type Node struct {
	//public
	Description string
	MaxRetry    int
	NoPersist   bool
	Name        string
	//private
	uuid            string
	nbretry         int
	stackOnPush     bool
	nextTriggerTime time.Time
}

//MarshalBinary implement marshal
func (n Node) MarshalBinary() ([]byte, error) {

	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := n.NodeGobEncode(enc)
	return buf.Bytes(), err

}

//UnmarshalBinary Implement unmarshall
func (n *Node) UnmarshalBinary(data []byte) error {

	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := n.NodeGobDecode(dec)
	return err
}

//Init implement retryit interface
func (n Node) Init() {

}

//Execute implement retryit interface
func (n Node) Execute() error {
	return nil
}

//GetData implement retryit interface
func (n Node) GetData() Node {

	return n
}

//SetData implement retryit interface
func (n Node) SetData(value Node) NodeRetry {

	return value
}

//OnSuccess implement retryit interface
func (n Node) OnSuccess() error {

	return ErrRetryItNotImplemented
}

//OnFailed implement retryit interface
func (n Node) OnFailed(e error) error {

	return ErrRetryItNotImplemented
}

//OnFinalFailed implement retryit interface
func (n Node) OnFinalFailed(e error) error {

	return ErrRetryItNotImplemented
}

//TimeToSleepOnNextFailed implement retryit interface
func (n Node) TimeToSleepOnNextFailed(retry int) time.Duration {

	return time.Second * time.Duration(1)
}

//GetAttempt number of retry
func (n *Node) GetAttempt() int {
	return n.nbretry
}

//NodeGobEncode Encode for serialise
func (n *Node) NodeGobEncode(enc *gob.Encoder) error {

	var err error
	err = enc.Encode(n.Description)
	if err != nil {
		return err
	}
	err = enc.Encode(n.MaxRetry)
	if err != nil {
		return err
	}
	err = enc.Encode(n.Name)
	if err != nil {
		return err
	}

	err = enc.Encode(n.uuid)
	if err != nil {
		return err
	}
	err = enc.Encode(n.nbretry)
	if err != nil {
		return err
	}

	err = enc.Encode(n.nextTriggerTime)
	if err != nil {
		return err
	}

	err = enc.Encode(n.NoPersist)
	if err != nil {
		return err
	}

	return nil
}

//NodeGobDecode Decode for deserialised
func (n *Node) NodeGobDecode(dec *gob.Decoder) error {

	var err error
	err = dec.Decode(&n.Description)
	if err != nil {
		return err
	}
	err = dec.Decode(&n.MaxRetry)
	if err != nil {
		return err
	}
	err = dec.Decode(&n.Name)
	if err != nil {
		return err
	}
	err = dec.Decode(&n.uuid)
	if err != nil {
		return err
	}
	err = dec.Decode(&n.nbretry)
	if err != nil {
		return err
	}

	err = dec.Decode(&n.nextTriggerTime)
	if err != nil {
		return err
	}

	err = dec.Decode(&n.NoPersist)
	if err != nil {
		return err
	}
	return nil
}

//Register Register default for type
func (rq *RetryQueue) Register(i interface{}, success func(interface{}) error, failed func(interface{}, error), finalFailed func(interface{}, error)) {
	if i != nil {
		gob.Register(i)
		i.(NodeRetry).Init()

		if success != nil {
			rq.registerSuccessFunc[reflect.ValueOf(i).Type().String()] = success
		}
		if failed != nil {
			rq.registerFailedFunc[reflect.ValueOf(i).Type().String()] = failed
		}
		if finalFailed != nil {
			rq.registerFinalFailedFunc[reflect.ValueOf(i).Type().String()] = finalFailed
		}
	}

}

//loadPersist Load the persist data
func (rq *RetryQueue) loadPersist() {
	rq.store.ParseAll(func(key string, data []byte) error {

		buffer := bytes.Buffer{}
		buffer.Write(data)
		dataDecoder := gob.NewDecoder(&buffer)
		var data2dec = make(map[string]interface{}, 1)
		var err error
		if err = dataDecoder.Decode(&data2dec); err == nil {
			for _, element := range data2dec {
				rq.EnqueueExecAndRetry(element)
			}

			return nil
		}
		return err

	})

}

//Init init queue and store
func (rq *RetryQueue) Init(store RStore) {
	var wg sync.WaitGroup
	rq.wg = &wg
	rq.waitingNodeList = list.New()
	rq.io = make(chan interface{})
	rq.registerSuccessFunc = make(map[string]func(interface{}) error)
	rq.registerFailedFunc = make(map[string]func(interface{}, error))
	rq.registerFinalFailedFunc = make(map[string]func(interface{}, error))
	rq.store = store
	rq.latency = 100

}

//Start Load in store and run loop
func (rq *RetryQueue) Start() {

	rq.loadPersist()

	go rq.runLoop()
}

func (n Node) String() string {
	return fmt.Sprintf("%s: retry_attempt:%d [%s]", n.Description, n.nbretry, n.uuid)
}

//RemoveByName Remove a node by name
func (rq *RetryQueue) RemoveByName(name string) {
	rq.EnqueueExecAndRetry(NodeRemove{name: name})
}

func (rq *RetryQueue) removeFromPersist(uuid string) error {
	return rq.Persist(uuid, nil)
}

//Persist save all queue on storage
func (rq *RetryQueue) Persist(uuid string, data interface{}) error {

	if data == nil {
		return rq.store.Delete(uuid)
	}

	return rq.store.Store(uuid, func() (*bytes.Buffer, error) {

		b := bytes.Buffer{}
		v := reflect.ValueOf(data)
		var data2enc = make(map[string]interface{}, 1)

		data2enc[v.Type().String()] = data
		dataEncoder := gob.NewEncoder(&b)
		if err := dataEncoder.Encode(data2enc); err == nil {
			return &b, nil
		}
		v = reflect.ValueOf(data)
		return nil, fmt.Errorf("Unable to persist this %s struct please try to register", v.Type())

	})

}

func (rq *RetryQueue) addNodeToWaitingList(data interface{}) {

	if rq.waitingNodeList.Front() == nil {
		rq.waitingNodeList.PushFront(data)
	} else {
		for e := rq.waitingNodeList.Front(); e != nil; e = e.Next() {
			if e.Next() == nil {
				rq.waitingNodeList.InsertAfter(data, e)
				break
			}
			if (data).(NodeRetry).GetData().nextTriggerTime.UnixNano() > ((e.Value).(NodeRetry)).GetData().nextTriggerTime.UnixNano() &&
				(data).(NodeRetry).GetData().nextTriggerTime.UnixNano() < ((e.Value).(NodeRetry)).GetData().nextTriggerTime.UnixNano() {
				rq.waitingNodeList.InsertAfter(data, e)
				break
			}
		}

	}

}

func (rq *RetryQueue) dataRetryNextAttempt(data NodeRetry, err error) {

	if returnSuccess := data.(NodeRetry).OnFailed(err); errors.Is(returnSuccess, ErrRetryItNotImplemented) {
		if failedFunc, ok := rq.registerFailedFunc[reflect.ValueOf(data).Type().String()]; ok {
			failedFunc(data, err)
		}
	}
	d := data.(NodeRetry).GetData()

	if d.nbretry == (d.MaxRetry - 1) {
		rq.removeFromPersist(d.uuid)
		if returnSuccess := data.(NodeRetry).OnFinalFailed(err); errors.Is(returnSuccess, ErrRetryItNotImplemented) {
			if finalFailedFunc, ok := rq.registerFinalFailedFunc[reflect.ValueOf(data).Type().String()]; ok {
				finalFailedFunc(data, err)
			}
		}

	} else {

		t := time.Now()
		d.nextTriggerTime = t.Add(data.(NodeRetry).TimeToSleepOnNextFailed(d.nbretry))
		d.stackOnPush = true
		if d.MaxRetry != MAXRETRYINFINITE {
			d.nbretry = d.nbretry + 1
		}

		data = data.(NodeRetry).SetData(d)
		rq.EnqueueExecAndRetry(data)
	}

}

func (rq *RetryQueue) runLoop() {

loop:
	for {
		select {
		case <-rq.stop:
			break loop
		case data, ok := <-rq.io:
			if ok {
				switch data.(type) {
				case delayedNode:

					t := time.Now()
					d := data.(delayedNode).node.(NodeRetry).GetData()
					d.nextTriggerTime = t.Add((data.(delayedNode).time))
					d.stackOnPush = true
					rq.EnqueueExecAndRetry(data.(delayedNode).node.(NodeRetry).SetData(d))

				case NodeRemove:
					for e := rq.waitingNodeList.Front(); e != nil; e = e.Next() {

						if e.Value.(NodeRetry).GetData().Name != data.(NodeRemove).name {
							rq.waitingNodeList.Remove(e)
							break
						}

					}
				case NodeRetry:
					d := data.(NodeRetry).GetData()
					if d.uuid == "" {
						d.uuid = xid.New().String()
						data = data.(NodeRetry).SetData(d)
					}

					if !d.NoPersist {
						rq.Persist(d.uuid, data)
					}

					if d.stackOnPush {
						d.stackOnPush = false
						data = data.(NodeRetry).SetData(d)
						rq.addNodeToWaitingList(data)
					} else {
						rq.wg.Add(1)

						go func() {
							defer rq.wg.Done()

							if err := data.(NodeRetry).Execute(); err == nil {
								var retryErr error
								removeNodeAfterSuccess := true
								if returnSuccess := data.(NodeRetry).OnSuccess(); returnSuccess != nil {
									retryErr = returnSuccess
									removeNodeAfterSuccess = false
									if errors.Is(returnSuccess, ErrRetryItNotImplemented) {
										if successFunc, ok := rq.registerSuccessFunc[reflect.ValueOf(data).Type().String()]; ok {
											if err := successFunc(data); err == nil {
												removeNodeAfterSuccess = true

											} else {
												retryErr = err
											}
										} else {
											removeNodeAfterSuccess = true
										}
									}
								}

								if removeNodeAfterSuccess {
									rq.removeFromPersist(d.uuid)

								} else {

									rq.dataRetryNextAttempt(data.(NodeRetry), retryErr)
								}

							} else {

								rq.dataRetryNextAttempt(data.(NodeRetry), err)
							}

						}()
					}

				default:
					fmt.Printf("Unknown type message receive %t\n", data)
				}

			}
		case <-time.After(time.Duration(rq.latency) * time.Millisecond):
			now := time.Now()

			for e := rq.waitingNodeList.Front(); e != nil; e = e.Next() {

				if now.UnixNano() > e.Value.(NodeRetry).GetData().nextTriggerTime.UnixNano() {
					rq.waitingNodeList.Remove(e)
					rq.EnqueueExecAndRetry(e.Value)
				}
			}
		}

	}

	rq.wg.Wait()

}

//EnqueueExecAndRetry Enqueue node and exec
func (rq *RetryQueue) EnqueueExecAndRetry(n interface{}) {

	go func() {
		rq.io <- n
	}()
}

//Stop Enqueue Stop the current queue
func (rq *RetryQueue) Stop() {
	rq.stop <- true
}
