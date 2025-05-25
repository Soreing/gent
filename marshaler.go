package gent

import (
	"encoding/json"
	"encoding/xml"
	"net/url"
)

// Marshaler defines how to process an object into byte array for a request's
// body, with additional optional headers to set.
type Marshaler func(body any) ([]byte, map[string][]string, error)

// JsonMarshaler uses the standard encoding/json marshaler to return the
// json encoded body and a Content-Type application/json header.
func JsonMarshaler(body any) (dat []byte, hdrs map[string][]string, err error) {
	hdrs = map[string][]string{"Content-Type": {"application/json"}}
	dat, err = json.Marshal(body)
	return
}

// XmlMarshaler uses the standard encoding/xml marshaler to return the
// xml encoded body and a Content-Type application/xml header.
func XmlMarshaler(body any) (dat []byte, hdrs map[string][]string, err error) {
	hdrs = map[string][]string{"Content-Type": {"application/xml"}}
	dat, err = xml.Marshal(body)
	return
}

// UrlEncodedMarshaler uses the standard net/url encoder to return the
// url encoded body and a Content-Type application/x-www-form-urlencoded header.
func UrlEncodedMarshaler(body any) (dat []byte, hdrs map[string][]string, err error) {
	if vals, ok := body.(map[string][]string); ok {
		dat = []byte(url.Values(vals).Encode())
	} else if vals, ok := body.(url.Values); ok {
		dat = []byte(vals.Encode())
	} else {
		return nil, nil, ErrInvalidBodyType
	}

	hdrs = map[string][]string{"Content-Type": {"application/x-www-form-urlencoded"}}
	return
}
