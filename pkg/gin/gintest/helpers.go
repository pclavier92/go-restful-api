package gintest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kylelemons/godebug/pretty"
	"github.com/pclavier92/go-restful-api/pkg/gin"
)

func init() {
	gin.SetTest()
}

// Fake is a simple test engine
type Fake struct {
	real *gin.Engine
}

// NewFake retours a test router
func NewFake() *Fake {
	return &Fake{gin.New("8080")}
}

func (f *Fake) toJSON(body io.ReadCloser, dest interface{}) error {
	// copy the body so we dont modify it!
	c := new(bytes.Buffer)
	_, err := io.Copy(c, body)
	if err != nil {
		return err
	}
	cc := c.Bytes()
	err = json.Unmarshal(cc, dest)
	return err
}

// GET will make a get request with the fake engine
func (f *Fake) GET(routerPath, path string, dest interface{}, cr gin.Controller) (*http.Response, error) {
	f.real.GET(routerPath, cr)
	w := httptest.NewRecorder()
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	f.real.ServeHTTP(w, req)
	res := w.Result()
	if dest != nil {
		err = f.toJSON(res.Body, dest)
	}
	return res, err
}

func (f *Fake) payloadRequest(path, method string, payload, dest interface{}, cr gin.Controller) (*http.Response, error) {
	w := httptest.NewRecorder()
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	r := bytes.NewReader(b)
	req, err := http.NewRequest(method, path, r)
	if err != nil {
		return nil, err
	}
	f.real.ServeHTTP(w, req)
	res := w.Result()
	if dest != nil {
		err = f.toJSON(res.Body, dest)
	}
	return res, err
}

// POST will make a POST request with the fake engine
func (f *Fake) POST(routerPath, path string, payload, dest interface{}, cr gin.Controller) (*http.Response, error) {
	f.real.POST(routerPath, cr)
	return f.payloadRequest(path, "POST", payload, dest, cr)
}

// PUT will make a PUT request with the fake engine
func (f *Fake) PUT(routerPath, path string, payload, dest interface{}, cr gin.Controller) (*http.Response, error) {
	f.real.PUT(routerPath, cr)
	return f.payloadRequest(path, "PUT", payload, dest, cr)
}

// PATCH will make a PATCH request with the fake engine
func (f *Fake) PATCH(routerPath, path string, payload, dest interface{}, cr gin.Controller) (*http.Response, error) {
	f.real.PATCH(routerPath, cr)
	return f.payloadRequest(path, "PATCH", payload, dest, cr)
}

func (f *Fake) Any(routerPath, path, method string, payload, dest interface{}, cr gin.Controller) (*http.Response, error) {
	f.real.Any(routerPath, cr)
	return f.payloadRequest(path, method, payload, dest, cr)
}

// Test is a simple struct to which you can GET, PUT or POST
// and will run a couple of basic checks and then return
// the response and the error of these operations
type Test struct {
	g      Fake
	called int
}

// NewTest returns a new Test to test eth API
func NewTest() *Test {
	return &Test{*NewFake(), 0}
}

// Wants describe what a Test method wants
type Wants struct {
	// JSON is the struct which the Test wants the endpoint to return
	JSON interface{}
	// Code is status code which we expect
	Code int
	// AnyErr will cancel all checks if ANY err occurred. Useful
	// when you just want to check if a function threw an error.
	AnyErr bool
	// Err is a SPECIFIC error which you want to see
	Err error
}

// When will set the situation for a Test method
type When struct {
	// Fmt is the desired path following gin's format convention for paths
	// Example: "/resource/:type"
	Fmt string
	// Path is the path with the parameters filled in
	// Example: "/resource/myType"
	Path string
	// Query is the query string
	// Example: "limit=10&isAwesome=true"
	Query string
	// Payload will be unmarhsaled and sent as the body
	// of the request
	Payload interface{}
	// Dest MUST BE a pointer. Test will unmarshal the JSON
	// response into this struct.
	Dest interface{}
}

func (Test) checks(t *testing.T, r *http.Response, err error, w Wants, d When) {
	t.Helper()
	if err != w.Err {
		if w.AnyErr {
			return
		}
		wanted := "nil"
		if w.Err != nil {
			wanted = w.Err.Error()
		}
		got := "nil"
		if err != nil {
			got = err.Error()
		}
		t.Errorf("Got Error! GOT: %s, WANTED: %s", got, wanted)
	}
	if w.Code != r.StatusCode {
		t.Errorf("Wrong code! GOT: %d, WANTED: %d", r.StatusCode, w.Code)
	}
	if d.Dest == nil {
		return
	}
	if diff := pretty.Compare(d.Dest, w.JSON); diff != "" {
		t.Errorf("Wrong JSON! Diff (-got, +want)\n%s", diff)
	}
}

func (tt *Test) request(t *testing.T, method string, d When, w Wants, cr gin.Controller) (*http.Response, error) {
	t.Helper()
	basePath := fmt.Sprintf("/%d", tt.called)
	ginPath := basePath + d.Fmt
	url := basePath + d.Path + d.Query
	var r *http.Response
	var err error
	fmt.Println(url)
	fmt.Println(ginPath)
	switch method {
	case "GET":
		r, err = tt.g.GET(ginPath, url, d.Dest, cr)
	case "POST":
		r, err = tt.g.POST(ginPath, url, d.Payload, d.Dest, cr)
	case "PUT":
		r, err = tt.g.PUT(ginPath, url, d.Payload, d.Dest, cr)
	case "PATCH":
		r, err = tt.g.PATCH(ginPath, url, d.Payload, d.Dest, cr)
	default:
		r, err = tt.g.Any(basePath+d.Fmt, method, url, d.Payload, d.Dest, cr)
	}
	tt.checks(t, r, err, w, d)
	tt.called++
	return r, err

}

// GET makes a GET request to a mock server. It will make the test fail if:
// * There was an err and w.AnyErr is false AND w.Err is not the error we got
// * There was not and error and w.AnyErr is true
// * The code is different from w.Code
// * The unmarhaled json is different from w.JSON
func (tt *Test) GET(t *testing.T, d When, w Wants, cr gin.Controller) (*http.Response, error) {
	t.Helper()
	return tt.request(t, "GET", d, w, cr)
}

// POST makes a POST request to a mock server. It will make the test fail if:
// * There was an err and w.AnyErr is false AND w.Err is not the error we got
// * There was not and error and w.AnyErr is true
// * The code is different from w.Code
// * The unmarhaled json is different from w.JSON
func (tt *Test) POST(t *testing.T, d When, w Wants, cr gin.Controller) (*http.Response, error) {
	t.Helper()
	return tt.request(t, "POST", d, w, cr)
}

// PUT makes a PUT request to a mock server. It will make the test fail if:
// * There was an err and w.AnyErr is false AND w.Err is not the error we got
// * There was not and error and w.AnyErr is true
// * The code is different from w.Code
// * The unmarhaled json is different from w.JSON
func (tt *Test) PUT(t *testing.T, d When, w Wants, cr gin.Controller) (*http.Response, error) {
	t.Helper()
	return tt.request(t, "PUT", d, w, cr)
}

// PATCH makes a PATCH request to a fake server. It will make the test fail if:
// * There was an err and w.AnyErr is false AND w.Err is not the error we got
// * There was not and error and w.AnyErr is true
// * The code is different from w.Code
// * The unmarhaled json is different from w.JSON
func (tt *Test) PATCH(t *testing.T, d When, w Wants, cr gin.Controller) (*http.Response, error) {
	t.Helper()
	return tt.request(t, "PATCH", d, w, cr)
}

// Any makes an arbitrary request to a fake server. The method must be specified in the parameters.
// It will make the test fail if:
// * There was an err and w.AnyErr is false AND w.Err is not the error we got
// * There was not and error and w.AnyErr is true
// * The code is different from w.Code
// * The unmarhaled json is different from w.JSON
func (tt *Test) Any(t *testing.T, method string, d When, w Wants, cr gin.Controller) (*http.Response, error) {
	t.Helper()
	return tt.request(t, method, d, w, cr)
}
