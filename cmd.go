package retry

import (
	"bytes"
	"context"
	"encoding/gob"
	"os/exec"
	"time"
)

//CMDSuccess CMDSuccess method
type CMDSuccess func(*CMD) error

//CMDFailed CMDFailed method
type CMDFailed func(*CMD, error)

//CMDFinalFailed CMDFinalFailed method
type CMDFinalFailed func(*CMD, error)

//CMD Struct
type CMD struct {
	Node
	CmdLine     string
	Args        []string
	WaitingTime int
}

//MarshalBinary implement retry interface
func (c CMD) MarshalBinary() ([]byte, error) {

	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := c.gobEncode(enc)
	return buf.Bytes(), err

}

func (c *CMD) gobEncode(enc *gob.Encoder) error {

	var err error
	err = c.Node.NodeGobEncode(enc)
	if err != nil {
		return err
	}
	err = enc.Encode(c.CmdLine)
	if err != nil {
		return err
	}
	err = enc.Encode(c.WaitingTime)
	if err != nil {
		return err
	}

	return nil
}

//UnmarshalBinary implement retry interface
func (c *CMD) UnmarshalBinary(data []byte) error {

	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := c.gobDecode(dec)
	return err
}

func (c *CMD) gobDecode(dec *gob.Decoder) error {

	var err error
	err = c.Node.NodeGobDecode(dec)
	if err != nil {
		return err
	}
	err = dec.Decode(&c.CmdLine)
	if err != nil {
		return err
	}

	err = dec.Decode(&c.WaitingTime)
	if err != nil {
		return err
	}
	return nil
}

//Init implement retry interface
func (c CMD) Init() {

}

//Execute implement retry interface
func (c CMD) Execute(ctx context.Context) error {
	var err error
	if err = exec.CommandContext(ctx, c.CmdLine, c.Args...).Run(); err != nil {
		// This will fail after 100 milliseconds. The 5 second sleep
		// will be interrupted.
	}

	return err
}

//GetData implement retry interface
func (c CMD) GetData() Node {

	return c.Node
}

//SetData implement retry interface
func (c CMD) SetData(n Node) NodeRetry {
	c.Node = n
	return c
}

//OnSuccess implement retry interface
func (c CMD) OnSuccess() error {

	return nil
}

//OnFailed implement retry interface
func (c CMD) OnFailed(e error) error {

	return ErrRetryNotImplemented
}

//OnFinalFailed implement retry interface
func (c CMD) OnFinalFailed(e error) error {

	return ErrRetryNotImplemented
}

//TimeToSleepOnNextFailed implement retry interface
func (c CMD) TimeToSleepOnNextFailed(retry int) time.Duration {

	return time.Second * time.Duration(c.WaitingTime)
}
