package gent

import (
	"fmt"
	"net/url"
	"testing"
)

// TestJsonMarshaler tests that the json marshaler can convert objects into
// byte arrays.
func TestJsonMarshaler(t *testing.T) {
	tests := []struct {
		Name  string
		Input any
		Data  []byte
		Type  string
		Error error
	}{
		{
			Name:  "Marshal nil",
			Input: nil,
			Data:  []byte("null"),
			Type:  "application/json",
			Error: nil,
		},
		{
			Name:  "Marshal value",
			Input: "200 Success",
			Data:  []byte(`"200 Success"`),
			Type:  "application/json",
			Error: nil,
		},
		{
			Name: "Marshal array",
			Input: []string{
				"123",
				"456",
				"789",
			},
			Data:  []byte(`["123","456","789"]`),
			Type:  "application/json",
			Error: nil,
		},
		{
			Name: "Marshal map/object",
			Input: map[string]any{
				"id":   123,
				"name": "John Smith",
			},
			Data:  []byte(`{"id":123,"name":"John Smith"}`),
			Type:  "application/json",
			Error: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			msh := NewJSONMarshaler()

			dat, ct, err := msh.Marshal(test.Input)

			if err != test.Error {
				t.Errorf("expected err to be %v but it's %v", test.Error, err)
			}
			if test.Data != nil && dat == nil {
				t.Errorf("expected data to not be nil")
			} else if string(dat) != string(test.Data) {
				t.Errorf("expected data to be %s but it's %s", string(test.Data), string(dat))
			}
			if ct != test.Type {
				t.Errorf("expected type to be %s but it's %s", string(test.Type), string(ct))
			}
		})
	}
}

// TestXmlMarshaler tests that the xml marshaler can convert objects into
// byte arrays.
func TestXmlMarshaler(t *testing.T) {
	tests := []struct {
		Name  string
		Input any
		Data  []byte
		Type  string
		Error error
	}{
		{
			Name:  "Marshal nil",
			Input: nil,
			Data:  nil,
			Type:  "application/xml",
			Error: nil,
		},
		{
			Name:  "Marshal value",
			Input: "200 Success",
			Data:  []byte(`<string>200 Success</string>`),
			Type:  "application/xml",
			Error: nil,
		},
		{
			Name: "Marshal array",
			Input: []string{
				"123",
				"456",
				"789",
			},
			Data:  []byte(`<string>123</string><string>456</string><string>789</string>`),
			Type:  "application/xml",
			Error: nil,
		},
		{
			Name: "Marshal map/object",
			Input: struct {
				XMLName string `xml:"employee"`
				Id      int    `xml:"id"`
				Name    string `xml:"name"`
			}{

				XMLName: "employee",
				Id:      123,
				Name:    "John Smith",
			},
			Data:  []byte(`<employee><id>123</id><name>John Smith</name></employee>`),
			Type:  "application/xml",
			Error: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			msh := NewXMLMarshaler()

			dat, ct, err := msh.Marshal(test.Input)

			if err != test.Error {
				t.Errorf("expected err to be %v but it's %v", test.Error, err)
			}
			if test.Data != nil && dat == nil {
				t.Errorf("expected data to not be nil")
			} else if string(dat) != string(test.Data) {
				t.Errorf("expected data to be %s but it's %s", string(test.Data), string(dat))
			}
			if ct != test.Type {
				t.Errorf("expected type to be %s but it's %s", string(test.Type), string(ct))
			}
		})
	}
}

// TestFormMarshaler tests that the form marshaler can convert url.Values into
// byte arrays.
func TestFormMarshaler(t *testing.T) {
	tests := []struct {
		Name  string
		Input any
		Data  []byte
		Type  string
		Error error
	}{
		{
			Name:  "Empty list",
			Input: url.Values{},
			Data:  []byte(""),
			Type:  "application/x-www-form-urlencoded",
			Error: nil,
		},
		{
			Name: "Populated list",
			Input: url.Values{
				"id":   {"123"},
				"name": {"John Smith"},
			},
			Data:  []byte(`id=123&name=John+Smith`),
			Type:  "application/x-www-form-urlencoded",
			Error: nil,
		},
		{
			Name:  "invalid type",
			Input: "invalid type",
			Data:  nil,
			Type:  "application/x-www-form-urlencoded",
			Error: fmt.Errorf("invalid body type"),
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			msh := NewFormMarshaler()

			dat, ct, err := msh.Marshal(test.Input)

			if test.Error != nil && err == nil {
				t.Errorf("expected err to not be nil")
			} else if test.Error == nil && err != nil {
				t.Errorf("expected err to be nil")
			} else if test.Error != nil && err.Error() != test.Error.Error() {
				t.Errorf("expected err to be %s but it's %s", test.Error.Error(), err.Error())
			}
			if test.Data != nil && dat == nil {
				t.Errorf("expected data to not be nil")
			} else if string(dat) != string(test.Data) {
				t.Errorf("expected data to be %s but it's %s", string(test.Data), string(dat))
			}
			if ct != test.Type {
				t.Errorf("expected type to be %s but it's %s", string(test.Type), string(ct))
			}
		})
	}
}
