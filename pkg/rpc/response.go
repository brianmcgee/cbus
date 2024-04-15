package rpc

import (
	"encoding/json"

	"github.com/juju/errors"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

type Response struct {
	Msg   *nats.Msg
	NKey  string
	Error error
	Value json.RawMessage
}

func (r Response) MarshalJSON() ([]byte, error) {
	var errorMsg string
	if r.Error != nil {
		errorMsg = r.Error.Error()
	}

	anon := struct {
		NKey  string          `json:"nkey"`
		Error string          `json:"error,omitempty"`
		Value json.RawMessage `json:"value,omitempty"`
	}{NKey: r.NKey, Error: errorMsg, Value: r.Value}

	return json.Marshal(anon)
}

func parseResponse(msg *nats.Msg, resp *Response) {
	resp.Msg = msg

	// todo use a constant
	resp.NKey = msg.Header.Get("NKey")
	if resp.NKey == "" {
		// todo make a const error
		resp.Error = errors.New("nkey header not found")
		return
	}

	err := msg.Header.Get(micro.ErrorHeader)
	errCode := msg.Header.Get(micro.ErrorCodeHeader)

	if !(err == "" || errCode == "") {
		resp.Error = errors.Errorf("%s: %s", errCode, err)
	} else {
		resp.Error = json.Unmarshal(msg.Data, &resp.Value)
	}
}
