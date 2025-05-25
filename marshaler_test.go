package gent

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestJsonMarshaler tests marshaling objects into JSON.
func TestJsonMarshaler(t *testing.T) {
	tests := []struct {
		Name    string
		Object  any
		Bytes   []byte
		Headers map[string][]string
		Error   error
	}{
		{
			Name:    "Marshal nil",
			Object:  nil,
			Bytes:   []byte("null"),
			Headers: map[string][]string{"Content-Type": {"application/json"}},
			Error:   nil,
		},
		{
			Name:    "Marshal value",
			Object:  "200 Success",
			Bytes:   []byte(`"200 Success"`),
			Headers: map[string][]string{"Content-Type": {"application/json"}},
			Error:   nil,
		},
		{
			Name: "Marshal array",
			Object: []string{
				"123",
				"456",
				"789",
			},
			Bytes:   []byte(`["123","456","789"]`),
			Headers: map[string][]string{"Content-Type": {"application/json"}},
			Error:   nil,
		},
		{
			Name: "Marshal map/object",
			Object: map[string]any{
				"id":   123,
				"name": "John Smith",
			},
			Bytes:   []byte(`{"id":123,"name":"John Smith"}`),
			Headers: map[string][]string{"Content-Type": {"application/json"}},
			Error:   nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			bts, hdrs, err := JsonMarshaler(test.Object)

			assert.Equal(t, test.Error, err)
			assert.Equal(t, test.Bytes, bts)
			assert.Equal(t, test.Headers, hdrs)
		})
	}
}

// TestXmlMarshaler tests marshaling objects into XML.
func TestXmlMarshaler(t *testing.T) {
	tests := []struct {
		Name    string
		Object  any
		Bytes   []byte
		Headers map[string][]string
		Error   error
	}{
		{
			Name:    "Marshal nil",
			Object:  nil,
			Bytes:   nil,
			Headers: map[string][]string{"Content-Type": {"application/xml"}},
			Error:   nil,
		},
		{
			Name:    "Marshal value",
			Object:  "200 Success",
			Bytes:   []byte(`<string>200 Success</string>`),
			Headers: map[string][]string{"Content-Type": {"application/xml"}},
			Error:   nil,
		},
		{
			Name: "Marshal array",
			Object: []string{
				"123",
				"456",
				"789",
			},
			Bytes:   []byte(`<string>123</string><string>456</string><string>789</string>`),
			Headers: map[string][]string{"Content-Type": {"application/xml"}},
			Error:   nil,
		},
		{
			Name: "Marshal map/object",
			Object: struct {
				XMLName string `xml:"employee"`
				Id      int    `xml:"id"`
				Name    string `xml:"name"`
			}{

				XMLName: "employee",
				Id:      123,
				Name:    "John Smith",
			},
			Bytes:   []byte(`<employee><id>123</id><name>John Smith</name></employee>`),
			Headers: map[string][]string{"Content-Type": {"application/xml"}},
			Error:   nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			bts, hdrs, err := XmlMarshaler(test.Object)

			assert.Equal(t, test.Error, err)
			assert.Equal(t, test.Bytes, bts)
			assert.Equal(t, test.Headers, hdrs)
		})
	}
}

// TestUrlEncodedMarshaler tests encoding url.Values objects into byte arrays.
func TestUrlEncodedMarshaler(t *testing.T) {
	tests := []struct {
		Name    string
		Values  any
		Bytes   []byte
		Headers map[string][]string
		Error   error
	}{
		{
			Name:    "Without values",
			Values:  map[string][]string{},
			Bytes:   []byte(""),
			Headers: map[string][]string{"Content-Type": {"application/x-www-form-urlencoded"}},
			Error:   nil,
		},
		{
			Name: "With values",
			Values: map[string][]string{
				"id":   {"123"},
				"name": {"John Smith"},
			},
			Bytes:   []byte(`id=123&name=John+Smith`),
			Headers: map[string][]string{"Content-Type": {"application/x-www-form-urlencoded"}},
			Error:   nil,
		},
		{
			Name: "As url.Values",
			Values: url.Values{
				"id":   {"123"},
				"name": {"John Smith"},
			},
			Bytes:   []byte(`id=123&name=John+Smith`),
			Headers: map[string][]string{"Content-Type": {"application/x-www-form-urlencoded"}},
			Error:   nil,
		},
		{
			Name:    "Invalid type",
			Values:  "string",
			Bytes:   nil,
			Headers: nil,
			Error:   ErrInvalidBodyType,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			bts, hdrs, err := UrlEncodedMarshaler(test.Values)

			assert.Equal(t, test.Error, err)
			assert.Equal(t, test.Bytes, bts)
			assert.Equal(t, test.Headers, hdrs)
		})
	}
}
