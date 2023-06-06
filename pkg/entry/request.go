package entry

import (
	"bytes"
	"errors"
	"io"
	"sync"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

var ErrBadJSON = errors.New("api: invalid json ")

// Request is sent by the vince script embedded in client websites
type Request struct {
	EventName   string `json:"n"`
	URI         string `json:"url"`
	Referrer    string `json:"r"`
	Domain      string `json:"d"`
	ScreenWidth int    `json:"w"`
	HashMode    bool   `json:"h"`

	// Used in tests
	IP        string `json:"ip,omitempty"`
	UserAgent string `json:"ua,omitempty"`

	b bytes.Buffer
}

// Parse opportunistic parses request body to r object. This is crucial method
// any gains here translates to smooth  events ingestion pipeline.
//
// A hard size limitation of 32kb is imposed. This is arbitrary value, any change
// to it must be be supported with statistics.
func (r *Request) Parse(body io.Reader) error {
	r.b.ReadFrom(io.LimitReader(body, 32<<10))
	return json.Unmarshal(r.b.Bytes(), r)
}

func (r *Request) Release() {
	r.b.Reset()
	requestPool.Put(r)
}

func NewRequest() *Request {
	return requestPool.Get().(*Request)
}

var requestPool = &sync.Pool{
	New: func() any {
		return new(Request)
	},
}
