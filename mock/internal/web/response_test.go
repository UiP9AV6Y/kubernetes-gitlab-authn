package web

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRespondError(t *testing.T) {
	tests := map[string]struct {
		haveError   string
		haveCode    int
		wantBody    string
		wantCode    int
		wantHeaders map[string]string
	}{
		"empty": {
			haveError: "",
			haveCode:  http.StatusInternalServerError,
			wantBody:  `{"error":""}`,
			wantCode:  http.StatusInternalServerError,
			wantHeaders: map[string]string{
				HeaderContentType: ContentTypeJSON,
			},
		},
		"error": {
			haveError: "test",
			haveCode:  http.StatusForbidden,
			wantBody:  `{"error":"test"}`,
			wantCode:  http.StatusForbidden,
			wantHeaders: map[string]string{
				HeaderContentType: ContentTypeJSON,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			respondError(w, test.haveCode, test.haveError)

			resp := w.Result()
			body, err := io.ReadAll(resp.Body)

			if err != nil {
				t.Fatalf("respondError(w, %d, %q) produced unreadable body: %q", test.haveCode, test.haveError, err)
			}

			if resp.StatusCode != test.wantCode {
				t.Errorf("respondError(w, %d, %q) produced status code %d; expected %d", test.haveCode, test.haveError, resp.StatusCode, test.wantCode)
			}

			if got, want := string(body), test.wantBody+"\n"; got != want {
				t.Errorf("respondError(w, %d, %q) produced body %q; expected %q", test.haveCode, test.haveError, got, want)
			}

			for h, want := range test.wantHeaders {
				if got := resp.Header.Get(h); got != want {
					t.Errorf("respondError(w, %d, %q) produced header %q with %q; expected %q", test.haveCode, test.haveError, h, got, want)
				}
			}
		})
	}
}

func TestRespondDTO(t *testing.T) {
	tests := map[string]struct {
		haveDTO     interface{}
		wantBody    string
		wantCode    int
		wantHeaders map[string]string
	}{
		"empty_string": {
			haveDTO:  "",
			wantBody: `""`,
			wantCode: http.StatusOK,
			wantHeaders: map[string]string{
				HeaderContentType: ContentTypeJSON,
			},
		},
		"empty_array": {
			haveDTO:  []string{},
			wantBody: `[]`,
			wantCode: http.StatusOK,
			wantHeaders: map[string]string{
				HeaderContentType: ContentTypeJSON,
			},
		},
		"empty_map": {
			haveDTO:  map[string]string{},
			wantBody: `{}`,
			wantCode: http.StatusOK,
			wantHeaders: map[string]string{
				HeaderContentType: ContentTypeJSON,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			w := httptest.NewRecorder()

			respondDTO(w, test.haveDTO)

			resp := w.Result()
			body, err := io.ReadAll(resp.Body)

			if err != nil {
				t.Fatalf("respondDTO(w, %v) produced unreadable body: %q", test.haveDTO, err)
			}

			if resp.StatusCode != test.wantCode {
				t.Errorf("respondDTO(w, %v) produced status code %d; expected %d", test.haveDTO, resp.StatusCode, test.wantCode)
			}

			if got, want := string(body), test.wantBody+"\n"; got != want {
				t.Errorf("respondDTO(w, %v) produced body %q; expected %q", test.haveDTO, got, want)
			}

			for h, want := range test.wantHeaders {
				if got := resp.Header.Get(h); got != want {
					t.Errorf("respondDTO(w, %v) produced header %q with %q; expected %q", test.haveDTO, h, got, want)
				}
			}
		})
	}
}
