---
title: Trace Customization
weight: 20
prev: /docs/features
next: /docs/built-in
---

Orchestrion offers serveral ways to control the traces produced by instrumented
applications.

## Prevent instrumentation of a section of code

By default, `orchestrion` injects instrumentation _everywhere_ possible. This
ensures users get the maximum possible coverage from their applications, as it
removes the possibility of someone forgetting to instrument a particular call.

There are however cases where you may want specific sections of your application
to not be instrumented, either because they result in excessively verbose
traces, or because those trace spans would be duplicated (for example, when
using a custom-made `net/http` middleware stack).

The `//dd:ignore` directive can be added anywhere in your application's code,
and will disable all `orchestrion` instrumentation in the annotated syntax tree.

For example:

```go
package demo

import "net/http"

//dd:ignore I don't want any of this to be instrumented, ever.
func noInstrumentationThere() {
  // Orchestrion will never add or modify any code in this function
  // ... etc ...
}

func definitelyInstrumented() {
  // Orchestrion may add or modify code in this function
  // ... etc ...

  //dd:ignore This particular server will NOT be instrumented
  server := &http.Server {
    Addr:    "127.0.0.1:8080",
    Handler: internalServerHandler,
  }

  // Orchestrion may add or modify code further down in this function
  // ... etc ...
}
```

{{<callout emoji="⚠️">}}
In certain cases, `orchestrion` adds instrumentation on the library side
(sometimes referred to as _callee_ instrumentation; as opposed to _call site_
instrumentation).

In such cases, it is currently not possible to opt-out of instrumentation. This
is the case for:
- `net/http` client instrumentation
- `github.com/gorilla/mux` middleware instrumentation
{{</callout>}}

## Creating custom trace spans

Any function annotated with the `//dd:span` directive will result in a trace
span being created around the function's execution. The directive can optionally
provide custom span tags as `key:value` pairs (all parsed as literal strings):

```go
//dd:span tag-name:for other-tag:bar
func tracedFunction() {
  // This function will be represented as a span named "tracedFunction"
}
```

### Result Capture

Functions annotated with `//dd:span` which return an `error` value will
automatically annotate spans with the returned `error` information if that is
non-`nil`.

```go
package demo

import "errors"

//dd:span
func failableFunction() (any, error) {
  // This span will have error information attached automatically.
  return nil, errors.ErrUnsupported
}
```

### Operation Name

The name of the operation (span name) is determined using the following
precedence list (first non-empty is selected):

- The `span.name` tag specified as a directive argument
  ```go
  //dd:span span.name:operationName
  func tracedFunction() {
    // This function will be represented as a span named "operationName"
  }
  ```
- The name of the function (closures do not have a name)
  ```go
  //dd:span tag-name:for other-tag:bar
  func tracedFunction() {
    // This function will be represented as a span named "tracedFunction"
  }
  ```
- The value of the very first tag from the directive arguments list
  ```go
  //dd:span tag-name:spanName other-tag:bar
  tracedFunction := func() {
    // This function will be represented as a span named "spanName"
  }
  ```

### Trace Context Propagation

If the annotated function accepts a {{<godoc "context" "Context" >}} argument,
that context will be used for trace propagation. Otherwise, if the function
accepts a {{<godoc "net/http" "Request" "*">}} argument, the request's context
will be used for trace propagation.

Functions that accept neither solely rely on _goroutine local storage_ for trace
propagation. This means that traces may be split on _goroutine_ boundaries
unless a {{<godoc "context" "Context" >}} or {{<godoc "net/http" "Request" "*">}}
value carying trace context is passed across.

Trace context carrying {{<godoc "context" "Context" >}} values are those that:

- have been received by a `//dd:span` annotated function, as instrumentation
  will create a new trace root span if if did not already carry trace context
- are returned by:
  - {{<godoc "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer" "StartSpanFromContext" >}}
  - {{<godoc "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer" "ContextWithSpan" >}}

```go
package demo

//dd:span
func caller(ctx context.Context) {
  wait := make(chan struct{}, 1)
  defer close(wait)

  // Weaving the span context into the child goroutine
  go callee(ctx, wait)
  <-wait
}

//dd:span
func callee(ctx context.Context, done chan<- struct{}) {
  done <- struct{}{}
}
```

### Manual Instrumentation

The {{<godoc "gopkg.in/DataDog/dd-trace-go.v1">}} library can be used to
manually instrument sections of your code even when building with `orchestrion`.

You can use APIs such as {{<godoc "gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer" "StartSpanFromContext" >}}
to create spans in any section of your code. This can be useful to delimit a
specific section of your code with a span without having to refactor it in a
separate function (which would allow the use of the `//dd:span` directive), or
when you need to customize the span more than the `//dd:span` directive allows.

{{<callout emoji="⚠️">}}
You may also use integrations from the packages within
{{<godoc "gopkg.in/DataDog/dd-trace-go.v1/contrib">}}, although this may result
in duplicated trace spans if `orchestrion` supports automatic instrumentation of
the same integration.

This can be useful to instrument calls that `orchestrion` does not yet support.
If you directly use integrations, we encourage you carefully review the
[release notes](https://github.com/DataDog/orchestrion/releases) before
upgrading to a new `orchestrion` release, so you can remove manual
instrumentation that was made redundant as necessary.
{{</callout>}}