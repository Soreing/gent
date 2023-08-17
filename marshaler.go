package gent

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/url"
)

// Marshaler  defines how to convert an object into a byte  array and its
// content type for making HTTP requests.
type Marshaler interface {
	Marshal(body any) (data []byte, content string, err error)
}

// jsonMarshaler is a marshaler for application/json content type.
type jsonMarshaler struct{}

// NewJSONMarshaler creates a marshaler for application/json content type.
func NewJSONMarshaler() Marshaler {
	return &jsonMarshaler{}
}

// Marshal returns an object encoded into a byte array and its content type.
func (m *jsonMarshaler) Marshal(v any) (dat []byte, t string, err error) {
	t = "application/json"
	dat, err = json.Marshal(v)
	return
}

// xmlMarshaler is a marshaler for application/xml content type.
type xmlMarshaler struct{}

// NewXMLMarshaler creates a marshaler for application/xml content type.
func NewXMLMarshaler() Marshaler {
	return &xmlMarshaler{}
}

// Marshal returns an object encoded into a byte array and its content type.
func (m *xmlMarshaler) Marshal(v any) (dat []byte, t string, err error) {
	t = "application/xml"
	dat, err = xml.Marshal(v)
	return
}

// formMarshaler is a marshaler for application/x-www-form-urlencoded content type.
type formMarshaler struct{}

// NewXMLMarshaler creates a marshaler for application/x-www-form-urlencoded content type.
func NewFormMarshaler() Marshaler {
	return &formMarshaler{}
}

// Marshal returns an object encoded into a byte array and its content type.
func (m *formMarshaler) Marshal(v any) (dat []byte, t string, err error) {
	t = "application/x-www-form-urlencoded"
	if fields, ok := v.(url.Values); !ok {
		err = fmt.Errorf("invalid body type")
	} else {
		dat = []byte(fields.Encode())
	}
	return
}
