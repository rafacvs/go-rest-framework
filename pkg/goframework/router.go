package goframework

import (
	"errors"
	"net/url"
	"strings"
)

type route struct {
	method   string
	pattern  string
	segments []routeSegment
	handler  HandlerFunc
}

type routeSegment struct {
	value    string
	isParam  bool
	paramKey string
}

type Router struct {
	routes []route
}

func NewRouter() *Router {
	return &Router{routes: make([]route, 0)}
}

func (r *Router) AddRoute(method, pattern string, handler HandlerFunc) error {
	if method == "" {
		return errors.New("router: http method vazio")
	}
	if pattern == "" {
		return errors.New("router: pattern vazio")
	}
	if handler == nil {
		return errors.New("router: handler nil")
	}

	method = strings.ToUpper(method)
	pattern = normalizePattern(pattern)

	for _, existing := range r.routes {
		if existing.method == method && existing.pattern == pattern {
			return errors.New("router: rota duplicada para " + method + " " + pattern)
		}
	}

	segments, err := buildSegments(pattern)
	if err != nil {
		return err
	}

	r.routes = append(r.routes, route{
		method:   method,
		pattern:  pattern,
		segments: segments,
		handler:  handler,
	})
	return nil
}

func (r *Router) Match(method, path string) (HandlerFunc, map[string]string, bool) {
	if method == "" {
		return nil, nil, false
	}

	method = strings.ToUpper(method)
	pathSegments := splitPath(path)

	for _, rt := range r.routes {
		if rt.method != method {
			continue
		}

		params, ok := matchSegments(rt.segments, pathSegments)
		if !ok {
			continue
		}
		return rt.handler, params, true
	}
	return nil, nil, false
}

func matchSegments(patternSegs, pathSegs []routeSegment) (map[string]string, bool) {
	if len(patternSegs) != len(pathSegs) {
		return nil, false
	}

	params := make(map[string]string)
	for idx, seg := range patternSegs {
		pathSeg := pathSegs[idx]

		if seg.isParam {
			value, err := url.PathUnescape(pathSeg.value)
			if err != nil {
				return nil, false
			}
			params[seg.paramKey] = value
			continue
		}

		if seg.value != pathSeg.value {
			return nil, false
		}
	}

	return params, true
}

func buildSegments(pattern string) ([]routeSegment, error) {
	segments := splitPath(pattern)
	result := make([]routeSegment, 0, len(segments))

	for _, seg := range segments {
		if strings.HasPrefix(seg.value, ":") {
			param := strings.TrimPrefix(seg.value, ":")
			if param == "" {
				return nil, errors.New("router: nome de parametro inválido em " + pattern)
			}
			result = append(result, routeSegment{
				value:    seg.value,
				isParam:  true,
				paramKey: param,
			})
			continue
		}

		result = append(result, seg)
	}

	return result, nil
}

func splitPath(path string) []routeSegment {
	clean := normalizePattern(path)
	if clean == "/" {
		return []routeSegment{}
	}

	rawSegments := strings.Split(strings.TrimPrefix(clean, "/"), "/")
	segments := make([]routeSegment, 0, len(rawSegments))
	for _, seg := range rawSegments {
		segments = append(segments, routeSegment{value: seg})
	}
	return segments
}

func normalizePattern(pattern string) string {
	if pattern == "" {
		return "/"
	}
	clean := strings.TrimSpace(pattern)
	if clean == "" {
		return "/"
	}
	if !strings.HasPrefix(clean, "/") {
		clean = "/" + clean
	}
	clean = strings.TrimSuffix(clean, "/")
	if clean == "" {
		return "/"
	}
	return clean
}
