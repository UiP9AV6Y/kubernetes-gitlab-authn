package handler_test

import (
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/UiP9AV6Y/kubernetes-gitlab-authn/pkg/handler"
)

func TestRequestIdentifier(t *testing.T) {
	sink := func(w http.ResponseWriter, r *http.Request) {
		// write the calculated value into the response for later verification
		io.WriteString(w, r.Header.Get(handler.HeaderRequestId))
	}
	// use large value to ease the testing of overflows
	//
	// using a static value enables us to test the scenario
	// where multiple requests are served at the same time
	sut := handler.LinearRequestIdentifier(uint64(math.MaxUint32), http.HandlerFunc(sink))

	// start at 2 as the counter is initialized with the time once and
	// we test the calculation result, at which point now() has been called twice
	//
	// we run the tests a few times to ensure integer overflows are handled gracefully
	for i := 2; i <= 11; i++ {
		r := httptest.NewRequest(http.MethodGet, "http://example.com/", nil)
		w := httptest.NewRecorder()

		sut.ServeHTTP(w, r)

		resp := w.Result()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Error("unable to read response body", err)
		}

		want := strconv.FormatUint(uint64(math.MaxUint32)*uint64(i), 10)
		if got := string(body); got != want {
			t.Errorf("RequestIdentifier(%d) = %q; want %q", i, got, want)
		}
	}
}
