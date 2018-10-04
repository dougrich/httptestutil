package httptestutil

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type ResponseAssertion func (*testing.T, *httptest.ResponseRecorder)

type RequestModifier func (req *http.Request)

type TestConfig struct {
	name string
	method string
	route string
	body string
	modifiers []RequestModifier
	assertions []ResponseAssertion
}

type TestOption func (*TestConfig)

type TestSet []TestConfig

// TestSet.Run runs the tests against a given handler inside the greater testing context. It runs each test as a sub test.
func (tests TestSet) Run(t *testing.T, handler http.Handler) {
	for _, test := range tests {
		t.Run(test.name, func (t *testing.T) {
			
			req, err := http.NewRequest(test.method, test.route, strings.NewReader(test.body))

			if err != nil {
				t.Fatal(err)
			}

			for _, modifier := range test.modifiers {
				modifier(req)
			}

			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, req)

			for _, assert := range test.assertions {
				assert(t, recorder)
			}
		})
	}
}

// Test creates a single test object
func Test(name string, options ...TestOption) TestConfig {
	test := TestConfig{
		name,
		"",
		"",
		"",
		[]RequestModifier{},
		[]ResponseAssertion{},
	}

	for _, option := range options {
		option(&test)
	}
	
	return test
}

/*
---

RequestModifiers

These modify the request going out

---
*/


// RequestMethod sets the method for the request
func RequestMethod(method string) TestOption {
	return func (test *TestConfig) {
		test.method = method
	}
}

/*
---

ResponseAssertions

These assertions validate properties of the response

For most cases you can use the responseAssertion helper following this template:

func <name>(<your arguments>) TestConfig {
	return responseAssertion(func (t *testing.T, recoreder *httptest.ResponseRecorder) {
		<your test>
	})
}

---
*/

// responseAssertion is a helper function that reduces boilerplate for creating HTTPTestAssertions
func responseAssertion(assertion ResponseAssertion) TestOption {
	return func (test *TestConfig) {
		test.assertions = append(test.assertions, assertion)
	}
}

// ResponseStatus asserts that the response received has the expected status
func ResponseStatus(expectedStatus int) TestOption {
	return responseAssertion(func (t *testing.T, recorder *httptest.ResponseRecorder) {
		if expectedStatus != recorder.Code {
			t.Errorf("Unexpected status code: recieved [%d], want [%d]", recorder.Code, expectedStatus)
		}
	})
}

// ResponseBody asserts that the response received has the expected body
func ResponseBody(expectedBody string) TestOption {
	return responseAssertion(func (t *testing.T, recorder *httptest.ResponseRecorder) {
		if responseBody := recorder.Body.String(); responseBody != expectedBody {
			t.Errorf("Unexpected body: received\n\n%v\n\nexpected\n\n%v", responseBody, expectedBody)
		}
	})
}