package web

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPaginationWriteHeader(t *testing.T) {
	haveMethod := http.MethodGet
	haveURL := "http://example.com/foo"
	tests := map[string]struct {
		haveSize    int
		havePage    int
		haveItems   int
		wantheaders map[string]string
	}{
		"default_empty": {
			haveSize:  DefaultBatchSize,
			havePage:  DefaultPage,
			haveItems: 0,
			wantheaders: map[string]string{
				HeaderNextPage:     "",
				HeaderCurrentPage:  "1",
				HeaderPreviousPage: "",
				HeaderTotalPages:   "0",
				HeaderPageSize:     "20",
				HeaderTotalItems:   "0",
				HeaderLink:         "",
			},
		},
		"default_page": {
			haveSize:  DefaultBatchSize,
			havePage:  DefaultPage,
			haveItems: DefaultBatchSize * 2,
			wantheaders: map[string]string{
				HeaderNextPage:     "2",
				HeaderCurrentPage:  "1",
				HeaderPreviousPage: "",
				HeaderTotalPages:   "2",
				HeaderPageSize:     "20",
				HeaderTotalItems:   "40",
				HeaderLink:         `<http://example.com/foo?page=2&per_page=20>; rel="next", <http://example.com/foo?page=2&per_page=20>; rel="last"`,
			},
		},
		"middle_page": {
			haveSize:  10,
			havePage:  5,
			haveItems: 100,
			wantheaders: map[string]string{
				HeaderNextPage:     "6",
				HeaderCurrentPage:  "5",
				HeaderPreviousPage: "4",
				HeaderTotalPages:   "10",
				HeaderPageSize:     "10",
				HeaderTotalItems:   "100",
				HeaderLink:         `<http://example.com/foo?page=4&per_page=10>; rel="prev", <http://example.com/foo?page=6&per_page=10>; rel="next", <http://example.com/foo?page=1&per_page=10>; rel="first", <http://example.com/foo?page=10&per_page=10>; rel="last"`,
			},
		},
		"last_page": {
			haveSize:  2,
			havePage:  4,
			haveItems: 8,
			wantheaders: map[string]string{
				HeaderNextPage:     "",
				HeaderCurrentPage:  "4",
				HeaderPreviousPage: "3",
				HeaderTotalPages:   "4",
				HeaderPageSize:     "2",
				HeaderTotalItems:   "8",
				HeaderLink:         `<http://example.com/foo?page=3&per_page=2>; rel="prev", <http://example.com/foo?page=1&per_page=2>; rel="first"`,
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			subject := NewPagination(test.haveSize, test.havePage, test.haveItems)
			req := httptest.NewRequest(haveMethod, haveURL, nil)
			w := httptest.NewRecorder()

			subject.WriteHeader(w, req)

			resp := w.Result()
			_, _ = io.ReadAll(resp.Body)

			for h, want := range test.wantheaders {
				if got := resp.Header.Get(h); got != want {
					t.Errorf("Pagination.WriteHeader(w, %q) produced header %q with %q; expected %q", haveURL, h, got, want)
				}
			}
		})
	}
}
