# HTTPTestUtil

Contains shared utilities for running http tests.

You craft a `httptestutil.TestSet`, or a set of assertions about the behavior of a handler. This consists of multiple `httptestutil.TestConfig`, which are created using the `httptestutil.Test` function - this takes a name for the test, and then a series of `httptestutil.TestOption` which either modify the request or create assertions about the response.

This is most useful for blackbox testing of a handler, where you don't understand the internals.

## Simple Example
```golang
package main

import (
	"testing"
	"net/http"
	"github.com/dougrich/httptestutil"
)

func TestHandler(t *testing.T) {
	httptestutil.TestSet{
		httptestutil.Test(
			"RejectGET405",
			httptestutil.RequestMethod(http.MethodGet),
			httptestutil.ResponseStatus(http.StatusMethodNotAllowed),
		),
	}.Run(t, &Handler{})
}
```