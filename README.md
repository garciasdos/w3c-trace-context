# w3c-trace-context

![License](https://img.shields.io/github/license/dschanoeh/w3c-trace-context)
[![Go Reference](https://pkg.go.dev/badge/github.com/dschanoeh/w3c-trace-context.svg)](https://pkg.go.dev/github.com/dschanoeh/w3c-trace-context)

This Go library implements the
[W3C Trace Context Specification](https://www.w3.org/TR/trace-context/).
It provides methods to parse existing Trace Context headers, to generate new
headers or to mutate existing headers in accordance with the specification.

## Usage

```go
package main

import . "github.com/dschanoeh/w3c-trace-context"

func main() {
    // Start with empty headers or pass existing input headers
    headers := http.Header{}

    // Define a parent ID for the operation or leave empty for a randomly
    // generated one
    parentId := "myOperation" 
    // Optionally add a new tracestate member
    member := TraceStateMember {
        Key: "myKey",
        Value: "myValue",
    }
    newHeaders, tc, err := HandleTraceContext(&headers, parentId, &member, SamplingBehaviorPassThrough)

    // pass newHeaders to next systems
}
```
