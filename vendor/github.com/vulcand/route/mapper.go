package route

import (
	"net/http"
	"strings"
)

// requestMapper maps the request to string e.g. maps request to it's hostname, or request to header
type requestMapper interface {
	// separator returns the separator that makes sense for this request, e.g. / for urls or . for domains
	separator() byte
	// equals returns the equivalent mapper if the two mappers are equivalent, e.g. map to the same sequence
	// mappers are also equivalent if one mapper is subset of another, e.g. combined mapper (host, path) is equivalent of (host) mapper
	equivalent(requestMapper) requestMapper
	// mapRequest maps request to string, e.g. request to it's URL path
	mapRequest(r *http.Request) string
	// newIter returns the iterator instead of string for stream matchers
	newIter(r *http.Request) *charIter
}

type methodMapper struct {
}

func (m *methodMapper) separator() byte {
	return methodSep
}

func (m *methodMapper) equivalent(o requestMapper) requestMapper {
	_, ok := o.(*methodMapper)
	if ok {
		return m
	}
	return nil
}

func (m *methodMapper) mapRequest(r *http.Request) string {
	return r.Method
}

func (m *methodMapper) newIter(r *http.Request) *charIter {
	return newIter([]string{m.mapRequest(r)}, []byte{m.separator()})
}

type pathMapper struct {
}

func (m *pathMapper) separator() byte {
	return pathSep
}

func (p *pathMapper) equivalent(o requestMapper) requestMapper {
	_, ok := o.(*pathMapper)
	if ok {
		return p
	}
	return nil
}

func (p *pathMapper) newIter(r *http.Request) *charIter {
	return newIter([]string{p.mapRequest(r)}, []byte{p.separator()})
}

func (p *pathMapper) mapRequest(r *http.Request) string {
	return rawPath(r)
}

type hostMapper struct {
}

func (p *hostMapper) equivalent(o requestMapper) requestMapper {
	_, ok := o.(*hostMapper)
	if ok {
		return p
	}
	return nil
}

func (m *hostMapper) separator() byte {
	return domainSep
}

func (h *hostMapper) mapRequest(r *http.Request) string {
	return strings.Split(strings.ToLower(r.Host), ":")[0]
}

func (p *hostMapper) newIter(r *http.Request) *charIter {
	return newIter([]string{p.mapRequest(r)}, []byte{p.separator()})
}

type headerMapper struct {
	header string
}

func (h *headerMapper) equivalent(o requestMapper) requestMapper {
	hm, ok := o.(*headerMapper)
	if ok && hm.header == h.header {
		return h
	}
	return nil
}

func (m *headerMapper) separator() byte {
	return headerSep
}

func (h *headerMapper) mapRequest(r *http.Request) string {
	return r.Header.Get(h.header)
}

func (h *headerMapper) newIter(r *http.Request) *charIter {
	return newIter([]string{h.mapRequest(r)}, []byte{h.separator()})
}

type seqMapper struct {
	seq []requestMapper
}

func newSeqMapper(seq ...requestMapper) *seqMapper {
	var out []requestMapper
	for _, s := range seq {
		switch m := s.(type) {
		case *seqMapper:
			out = append(out, m.seq...)
		default:
			out = append(out, s)
		}
	}
	return &seqMapper{seq: out}
}

func (s *seqMapper) newIter(r *http.Request) *charIter {
	out := make([]string, len(s.seq))
	for i := range s.seq {
		out[i] = s.seq[i].mapRequest(r)
	}
	seps := make([]byte, len(s.seq))
	for i := range s.seq {
		seps[i] = s.seq[i].separator()
	}
	return newIter(out, seps)
}

func (s *seqMapper) mapRequest(r *http.Request) string {
	out := make([]string, len(s.seq))
	for i := range s.seq {
		out[i] = s.seq[i].mapRequest(r)
	}
	return strings.Join(out, "")
}

func (s *seqMapper) separator() byte {
	return s.seq[0].separator()
}

func (s *seqMapper) equivalent(o requestMapper) requestMapper {
	so, ok := o.(*seqMapper)
	if !ok {
		return nil
	}

	var longer, shorter *seqMapper
	if len(s.seq) > len(so.seq) {
		longer = s
		shorter = so
	} else {
		longer = so
		shorter = s
	}
	for i, _ := range longer.seq {
		// shorter is subset of longer, return longer sequence mapper
		if i >= len(shorter.seq)-1 {
			return longer
		}
		if longer.seq[i].equivalent(shorter.seq[i]) == nil {
			return nil
		}
	}
	return longer
}

const (
	pathSep   = '/'
	domainSep = '.'
	headerSep = '/'
	methodSep = ' '
)
