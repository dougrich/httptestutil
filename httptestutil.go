package httptestutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
)

type Check func(*testing.T)

type ResponseAssertion func(*testing.T, *httptest.ResponseRecorder)

type RequestModifier func(req *http.Request)

type TestConfig struct {
	name       string
	method     string
	route      string
	body       string
	modifiers  []RequestModifier
	assertions []ResponseAssertion
	precheck	 []Check
	postcheck  []Check
}

type TestOption func(*TestConfig)

type TestSet []TestConfig

// TestSet.Run runs the tests against a given handler inside the greater testing context. It runs each test as a sub test.
func (tests TestSet) Run(t *testing.T, handler http.Handler) {
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			for _, check := range test.precheck {
				check(t)
			}

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

			for _, check := range test.postcheck {
				check(t)
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
		[]Check{},
		[]Check{},
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
	return func(test *TestConfig) {
		test.method = method
	}
}

// RequestHeader sets a specific header on the request
func RequestHeader(header string, value string) TestOption {
	return func(test *TestConfig) {
		test.modifiers = append(test.modifiers, func(req *http.Request) {
			req.Header.Set(header, value)
		})
	}
}

// RequestJSON sets the body and the correct content type for a request
func RequestJSON(d interface{}) TestOption {
	s, err := json.Marshal(d)
	if err != nil {
		// this is an error in test setup; panic here so that developers can quickly correct
		panic(err)
	}
	return func(test *TestConfig) {
		test.modifiers = append(test.modifiers, func(req *http.Request) {
			req.Header.Set("Content-Type", "application/json")
		})
		test.body = string(s)
	}
}

// RequestBody sets the body for a request
func RequestBody(body string) TestOption {
	return func(test *TestConfig) {
		test.body = body
	}
}

// RequestRel sets the relative url for the request (i.e. "/abc")
func RequestRel(rel string) TestOption {
	return func(test *TestConfig) {
		test.route = rel
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
	return func(test *TestConfig) {
		test.assertions = append(test.assertions, assertion)
	}
}

// ResponseStatus asserts that the response received has the expected status
func ResponseStatus(expectedStatus int) TestOption {
	return responseAssertion(func(t *testing.T, recorder *httptest.ResponseRecorder) {
		if expectedStatus != recorder.Code {
			t.Errorf("Unexpected status code: recieved [%d], want [%d]", recorder.Code, expectedStatus)
		}
	})
}

// ResponseBody asserts that the response received has the expected body
func ResponseBody(expectedBody string) TestOption {
	return responseAssertion(func(t *testing.T, recorder *httptest.ResponseRecorder) {
		if responseBody := recorder.Body.String(); responseBody != expectedBody {
			t.Errorf("Unexpected body: received\n\n%v\n\nexpected\n\n%v", responseBody, expectedBody)
		}
	})
}

// ResponseHeader asserts that the response received has the expected header
func ResponseHeader(header string, expected string) TestOption {
	return responseAssertion(func(t *testing.T, rr *httptest.ResponseRecorder) {
		if actual := rr.Header().Get(header); actual != expected {
			t.Errorf("Unexpected response header value for '%s': received '%s' expected '%s'", header, actual, expected)
		}
	})
}

// ResponseJsonField asserts that a specific JSON field has a specific value
func ResponseJsonField(field string, expected string) TestOption {
	return responseAssertion(func(t *testing.T, rr *httptest.ResponseRecorder) {
		responseBody := rr.Body.Bytes()

		var response map[string]interface{}
		if err := json.Unmarshal(responseBody, &response); err != nil {
			t.Errorf("Could not JSON parse response body: received\n\n%v", rr.Body)
		}

		if response[field] != expected {
			t.Errorf("Unexpected JSON field value for '%s': received '%s' expected '%s'", field, response[field], expected)
		}
	})
}

// ResponseJsonFieldPattern asserts that a specific JSON field matches a regular expression string
func ResponseJsonFieldPattern(field string, pattern string) TestOption {
	return responseAssertion(func(t *testing.T, rr *httptest.ResponseRecorder) {
		responseBody := rr.Body.Bytes()

		var response map[string]interface{}
		if err := json.Unmarshal(responseBody, &response); err != nil {
			t.Errorf("Could not JSON parse response body: received\n\n%v", rr.Body)
		}

		if actual, ok := response[field].(string); ok {

			if ok, _ := regexp.MatchString(pattern, actual); !ok {
				t.Errorf("Unexpected JSON field value for '%s': received '%s' expected pattern %s", field, actual, pattern)
			}
		} else {
			t.Errorf("Unexepected missing or non-string JSON field value for '%s'", field)
		}
	})
}

/*
---

Utility

These are utility functions that can be run before/after tests

---
*/

func Before(check Check) TestOption {
	return func (test *TestConfig) {
		test.precheck = append(test.precheck, check)
	}
}

func After(check Check) TestOption {
	return func (test *TestConfig) {
		test.precheck = append(test.postcheck, check)
	}
}