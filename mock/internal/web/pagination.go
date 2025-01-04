package web

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type Pagination struct {
	previous int
	current  int
	next     int
	size     int
	pages    int
	items    int

	previousValue string
	currentValue  string
	nextValue     string
	sizeValue     string
	pagesValue    string
	itemsValue    string
}

func NewPagination(size, page, items int) *Pagination {
	result := &Pagination{
		items:   items,
		size:    size,
		current: page,
	}
	pages := items / size
	result.pages = pages
	if mod := items % size; mod != 0 {
		result.pages += 1
	}

	if page <= 0 {
		result.next = page + 1
	} else if page < result.pages {
		result.previous = page - 1
		result.next = page + 1
	} else if page == result.pages {
		result.previous = page - 1
		result.next = 0
	}

	result.previousValue = strconv.Itoa(result.previous)
	result.currentValue = strconv.Itoa(result.current)
	result.nextValue = strconv.Itoa(result.next)
	result.sizeValue = strconv.Itoa(result.size)
	result.pagesValue = strconv.Itoa(result.pages)
	result.itemsValue = strconv.Itoa(result.items)

	return result
}

func (p *Pagination) Previous() int {
	return p.previous
}

func (p *Pagination) Current() int {
	return p.current
}

func (p *Pagination) Next() int {
	return p.next
}

func (p *Pagination) Size() int {
	return p.size
}

func (p *Pagination) Pages() int {
	return p.pages
}

func (p *Pagination) Items() int {
	return p.items
}

func (p *Pagination) RefLinks(base *url.URL) []string {
	links := make([]string, 0, 4)

	// copy URL and reset its query params
	link := *base
	link.RawQuery = ""
	link.Fragment = ""
	link.RawFragment = ""

	query := link.Query()
	query.Set(QueryParamSize, p.sizeValue)

	if p.previous > 0 {
		query.Set(QueryParamPage, p.previousValue)
		link.RawQuery = query.Encode()
		links = append(links, "<"+link.String()+">; rel=\"prev\"")
	}

	if p.next > 0 {
		query.Set(QueryParamPage, p.nextValue)
		link.RawQuery = query.Encode()
		links = append(links, "<"+link.String()+">; rel=\"next\"")
	}

	if p.current > 1 {
		query.Set(QueryParamPage, "1")
		link.RawQuery = query.Encode()
		links = append(links, "<"+link.String()+">; rel=\"first\"")
	}

	if p.current < p.pages {
		query.Set(QueryParamPage, p.pagesValue)
		link.RawQuery = query.Encode()
		links = append(links, "<"+link.String()+">; rel=\"last\"")
	}

	return links
}

func (p *Pagination) WriteHeader(w http.ResponseWriter, req *http.Request) {
	links := p.RefLinks(req.URL)

	w.Header().Set(HeaderCurrentPage, p.currentValue)
	w.Header().Set(HeaderTotalPages, p.pagesValue)
	w.Header().Set(HeaderTotalItems, p.itemsValue)
	w.Header().Set(HeaderPageSize, p.sizeValue)

	if p.previous > 0 {
		w.Header().Set(HeaderPreviousPage, p.previousValue)
	}

	if p.next > 0 {
		w.Header().Set(HeaderNextPage, p.nextValue)
	}

	if len(links) > 0 {
		w.Header().Set(HeaderLink, strings.Join(links, ", "))
	}
}
