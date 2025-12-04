package common

import "net/http"

type FilterHandle func(rw http.ResponseWriter, req *http.Request) error

type Filter struct {
	filterMap map[string]FilterHandle
}

func NewFilter() *Filter {
	return &Filter{filterMap: make(map[string]FilterHandle)}
}

func (f *Filter) RegisterFilterUri(uri string, handle FilterHandle) {
	f.filterMap[uri] = handle
}

func (f *Filter) GetFilterHandle(uri string) FilterHandle {
	return f.filterMap[uri]
}

type WebHandle func(rw http.ResponseWriter, req *http.Request)

func (f *Filter) Handle(webHandle WebHandle) func(rw http.ResponseWriter, req *http.Request) {
	return func(rw http.ResponseWriter, req *http.Request) {
		for path, handle := range f.filterMap {
			if path == req.RequestURI {
				err := handle(rw, req)
				if err != nil {
					rw.WriteHeader(500)
					return
				}
				break
			}
		}
		webHandle(rw, req)
	}
}
