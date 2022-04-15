# w3c-trace-context

This Go library implements the
[W3C Trace Context Specification](https://w3c.github.io/trace-context/).
It provides methods to parse existing Trace Context headers, to generate new
headers or to mutate existing headers in accordance with the specification.

## Usage

```go
func mutate() {
    // Start with empty headers or pass existing input headers
    headers := http.Header{}
	newHeaders, tc, err := HandleTraceContext(&headers, "vendor2", "val2", false)
    // pass newHeaders to next systems
}
```