# Go HTTP Client

![Build](https://github.com/soreing/gent/actions/workflows/build_status.yaml/badge.svg)
![Coverage](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/Soreing/4b6f950f01f3e6e5b9ed17b268664538/raw/gent)
[![Go Report Card](https://goreportcard.com/badge/github.com/Soreing/gent)](https://goreportcard.com/report/github.com/Soreing/gent)
[![Go Reference](https://pkg.go.dev/badge/github.com/Soreing/gent.svg)](https://pkg.go.dev/github.com/Soreing/gent)

Gent is a Golang HTTP Client wrapper that provides additional features for flexibility and increased performance.

## Usage

Create a Client that lets you to make requests. The constructor accepts a list of options to customize it.

```golang
// Default client
cl := gent.NewClient()
```

To make requests, use the functions named after HTTP methods. The following example sends a POST request to `http://localhost:8080/employees/create` with an application/json body containing an employee id and name, an Authorization header, and a query parameter set to true.

```golang
res, err := cl.Post(
    context.Background(),
    "http://localhost:8080/{}/{}",
    map[string]string{
        "id": "4481e035-1711-419f-82bc-bfb72da06375",
        "name": "John Smith",
    },
    gent.NewJSONMarshaler(),
    map[string]string{
 		"Authorization": "Bearer x.y.z",
	},
    map[string][]string{
 		"strict":    {"true"},
 	},
    "employees",
    "create",
)
```
### Request Builder ⭒NEW⭒ 
Requests can be constructed gradually with a request builder. Request require 
at least an HTTP method and an endpoint, the other parameters are optional.
```golang
res, err := cl.NewRequest(
    http.MethodGet, "http://localhost:8080/{}/{}",
).WithBody(
    map[string]string{
        "id": "4481e035-1711-419f-82bc-bfb72da06375",
        "name": "John Smith",
    },
    gent.NewJSONMarshaler(),
).WithHeaders(
    map[string]string{"Authorization": "Bearer x.y.z"},
).WithQueryParameters(
    map[string][]string{"strict": {"true"},},
).WithPathParameters(
    "employees", "create",
).Run(context.Background())
```


### Placeholders

The request's endpoint supports placeholders in the form of `{}`. Placeholders will be populated by the trailing variadic path parameters that get escaped before replacing the placeholders in the order they were provided.

### Request Body

Any data can be provided as a request body, and it is up to the marshaler to transform it into a byte array for the request. By default the package supports JSON, XML and Form marshalers. If you do not require a request body, leave both the body and the marshaler nil.

Marshalers must implement `Marshal(body any) (data []byte, content string, err error)`, which take any input and return a byte array data, a content type for the header and any error. The following is the implementation of the JSON Marshaler for reference.

```golang
type jsonMarshaler struct{}

func (m *jsonMarshaler) Marshal(
    v any,
) (dat []byte, t string, err error) {
	t = "application/json"
	dat, err = json.Marshal(v)
	return
}
```

### Query Parameters

Query parameters are constructed from a map and include a `?` if there is at least one query parameter provided. It is recommended to add query parameters via the map, as they get escaped. Parameters support arrays, which get added in the following format: 

```golang
map[string][]string{
    // ?ids=123&ids=456&ids=789
    "ids": {"123", "456", "789"}
}
```

## Options and Modules

The Client can be configured during creation with a variety of options.
```golang
cl := gent.NewClient(
    /* ... Option */
)
```

### HTTP Client

The client internally uses the default HTTP Client to make requests. This can be changed with the `UseHttpClient` option. The default behavior is to reuse clients between requests. This can also be changed by providing a constructor function that returns a new client for each request.

```golang
// Client that uses a new HTTP client
cl := gent.NewClient(
    gent.UseHttpClient(&http.Client{}),
)
```

```golang
// Client that creates a new HTTP client for each request
cl := gent.NewClient(
    gent.UseHttpClientConstructor(func() gent.HttpClient{
        return &http.Client{}
    }),
)
```

### Memory Pool

Each client uses a memory pool internally to speed up string operations by reusing memory allocations. By default, each client creates its own memory pool with default page sizes and pool sizes. A pre-configured memory pool can be provided to and even shared between clients.

```golang
cl := gent.NewClient(
    gent.UseMemoryPool(gent.NewMemPool(
        512, // Page size in bytes
        200, // Pool size in pages
    )),
)
```

You can provide your own implementation of memory pool if it implements how to acquire byte arrays from the pool with `Acquire() []byte` and release byte arrays into the pool with `Release(...[]byte)`

### Middlewares

Clients can be given middlewares that are executed for each request. Middlewares can be 
added to two stages. BeforeBuild middlewares run before the http.Request object is created,
while BeforeExecute middlewares run before the request is sent.

To add a middleware, use the Use function on the client.
```golang
cl := gent.NewClient()

// Middleware that adds an auth header to the request
cl.Use(
    gent.MDW_BeforeBuild, 
    func(c context.Context, r *gent.Request) {
        r.Headers["Authorization"] = "Bearer x.y.z"
        r.Next()
    },
)
```