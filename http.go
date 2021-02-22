package retry

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

//HTTPSuccess HTTPSuccess method
type HTTPSuccess func(*HTTP) error

//HTTPFailed HTTPFailed method
type HTTPFailed func(*HTTP, error)

//HTTPFinalFailed HTTPFinalFailed method
type HTTPFinalFailed func(*HTTP, error)

//HTTP Struct
type HTTP struct {
	Node
	URL             string
	Method          string
	HTTPPostData    map[string][]string
	HTTPGetData     map[string][]string
	HTTPHeader      map[string]string
	WaitingTime     int
	SuccessFunc     HTTPSuccess
	FailedFunc      HTTPFailed
	FinalFailedFunc HTTPFinalFailed
}

//MarshalBinary implement retryit interface
func (h HTTP) MarshalBinary() ([]byte, error) {

	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	err := h.gobEncode(enc)
	return buf.Bytes(), err

}

func (h *HTTP) gobEncode(enc *gob.Encoder) error {

	var err error
	err = h.Node.NodeGobEncode(enc)
	if err != nil {
		return err
	}
	err = enc.Encode(h.URL)
	if err != nil {
		return err
	}
	err = enc.Encode(h.Method)
	if err != nil {
		return err
	}
	err = enc.Encode(h.HTTPPostData)
	if err != nil {
		return err
	}
	err = enc.Encode(h.HTTPGetData)
	if err != nil {
		return err
	}
	err = enc.Encode(h.HTTPHeader)
	if err != nil {
		return err
	}
	err = enc.Encode(h.WaitingTime)
	if err != nil {
		return err
	}

	return nil
}

//UnmarshalBinary implement retryit interface
func (h *HTTP) UnmarshalBinary(data []byte) error {

	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	err := h.gobDecode(dec)
	return err
}

func (h *HTTP) gobDecode(dec *gob.Decoder) error {

	var err error
	err = h.Node.NodeGobDecode(dec)
	if err != nil {
		return err
	}
	err = dec.Decode(&h.URL)
	if err != nil {
		return err
	}
	err = dec.Decode(&h.Method)
	if err != nil {
		return err
	}
	err = dec.Decode(&h.HTTPPostData)
	if err != nil {
		return err
	}
	err = dec.Decode(&h.HTTPGetData)
	if err != nil {
		return err
	}
	err = dec.Decode(&h.HTTPHeader)
	if err != nil {
		return err
	}
	err = dec.Decode(&h.WaitingTime)
	if err != nil {
		return err
	}
	return nil
}

//Init implement retryit interface
func (h HTTP) Init() {

}

//Execute implement retryit interface
func (h HTTP) Execute() error {

	var postStrEncode = ""
	var getStrEncode = ""

	if h.HTTPPostData != nil {
		getStrEncode = url.Values(h.HTTPGetData).Encode()

	}

	if h.HTTPPostData != nil {
		postStrEncode = url.Values(h.HTTPPostData).Encode()

	}
	var u *url.URL
	var err error

	if u, err = url.ParseRequestURI(h.URL); err == nil {

		var r *http.Request

		u.RawQuery = getStrEncode
		client := &http.Client{}
		r, err = http.NewRequest(h.Method, u.String(), strings.NewReader(postStrEncode)) // URL-encoded payload
		if err != nil {
			return err
		}

		if h.HTTPHeader != nil {
			for headstr, headvalue := range h.HTTPHeader {
				r.Header.Set(headstr, headvalue)
			}
		}

		r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		r.Header.Add("Content-Length", strconv.Itoa(len(postStrEncode)))
		var resp *http.Response
		resp, err = client.Do(r)

		if err != nil {
			return err
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return fmt.Errorf("Request failed for %s (%s)", h.URL, resp.Status)
		}

		resp.Body.Close()

		return nil
	}

	return err

}

//GetData implement retryit interface
func (h HTTP) GetData() Node {

	return h.Node
}

//SetData implement retryit interface
func (h HTTP) SetData(n Node) NodeRetry {
	h.Node = n
	return h
}

//OnSuccess implement retryit interface
func (h HTTP) OnSuccess() error {
	if h.SuccessFunc != nil {
		return h.SuccessFunc(&h)

	}
	return ErrRetryNotImplemented
}

//OnFailed implement retryit interface
func (h HTTP) OnFailed(e error) error {
	if h.FailedFunc != nil {
		h.FailedFunc(&h, e)
		return nil
	}
	return ErrRetryNotImplemented
}

//OnFinalFailed implement retryit interface
func (h HTTP) OnFinalFailed(e error) error {
	if h.FinalFailedFunc != nil {
		h.FinalFailedFunc(&h, e)
		return nil
	}

	return ErrRetryNotImplemented
}

//TimeToSleepOnNextFailed implement retryit interface
func (h HTTP) TimeToSleepOnNextFailed(retry int) time.Duration {

	fmt.Printf("%s retry in %d seconds\n", h.URL, h.WaitingTime)
	return time.Second * time.Duration(h.WaitingTime)
}
