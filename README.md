# Go HTTP Client

![Build](https://github.com/soreing/gent/actions/workflows/build_status.yaml/badge.svg)
![Coverage](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/Soreing/4b6f950f01f3e6e5b9ed17b268664538/raw/gent)
[![Go Report Card](https://goreportcard.com/badge/github.com/Soreing/gent)](https://goreportcard.com/report/github.com/Soreing/gent)
[![Go Reference](https://pkg.go.dev/badge/github.com/Soreing/gent.svg)](https://pkg.go.dev/github.com/Soreing/gent)

Gent is a Golang HTTP Client wrapper that provides additional features.

## Usage

A Client allows you to make HTTP requests. Use NewDefaultClient to create one 
using http.DefaultClient, or NewClient to provide anything that implements the 
Requester interface. Client implements the same methods as the net/http Client 
to make requests.
```golang
//  type Requester interface {
//      Do(*http.Request) (*http.Response, error)
//      CloseIdleConnections()
//  }

cl := gent.NewClient(http.DefaultClient)

// simple get request
res, err := cl.Get("https://localhost:8080")
if err != nil {
    panic(err)
}

// complex patch request
body := bytes.NewBuffer([]byte(`{"name": "John Smith"}`))
req, err := http.NewRequest(http.MethodPatch, "https://localhost:8080", body)
if err != nil {
    panic(err)
}

req.Header.Set("Content-Type", "application/json")

res, err = cl.Do(req)
if err != nil {
    panic(err)
}
```

### Request Builder
A RequestBuilder assists in creating an *http.Request by providing a series of
chainable functions, placeholders and body marshaling.
```golang
obj := map[string]string{
    "id": "4481e035-1711-419f-82bc-bfb72da06375",
    "name": "John Smith",
}

// regular
rb := gent.NewRequest(http.MethodPost, "http://localhost:8080/{}/{}")
rb.WithBody(obj, gent.JsonMarshaler)
rb.WithHeader("Authorization", "Bearer x.y.z",)
rb.WithQueryParameter("strict", []string{"true"})
rb.WithPathParameters("employees", "create")
req, err := rb.Build(context.Background())

// chained
req, err := gent.NewRequest(
    http.MethodPost, "http://localhost:8080/{}/{}",
).WithBody(
    obj, gent.JsonMarshaler,
).WithHeader(
    "Authorization", "Bearer x.y.z",
).WithQueryParameter(
    "strict", []string{"true"},
).WithPathParameters(
    "employees", "create",
).Build(context.Background())
```

### Placeholders
The request's format supports placeholders in the form of `{}`. Placeholders 
will be replaced by encoded path parameters in the order they were provided.

### Request Body
Any object can be provided as a request body along with a marshaler that will
encode the object and attach some optional headers to the request. The package
provides JSON, XML and URL-Encoded Form marshalers, but any function with the 
signature `func(any) ([]byte, map[string][]string, error)` is a valid marshaler.

```golang
// JsonMarshaler uses the standard encoding/json marshaler to return the
// json encoded body and a Content-Type application/json header.
func JsonMarshaler(body any) (dat []byte, hdrs map[string][]string, err error) {
	hdrs = map[string][]string{"Content-Type": {"application/json"}}
	dat, err = json.Marshal(body)
	return
}
```

### Middlewares

A Client can use middleware-style functions that extend its behavior when making 
requests. Middleware functions run in the order they were attached, and the last
function performs the request. Each function has a request context to pass data
between middlewares, or to interact with the request or response. Use the Next()
method on the context to move to the next middleware. 
```golang
cl := gent.NewClient()

cl.Use(
    func(r *gent.Context) {
        now := time.Now()
        r.Next()
        fmt.Println("Elapsed time: ", time.Since(now))
    },
)
```