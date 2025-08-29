package httprouter

import (
	"context"
	"net/http"
	"regexp"
	"strings"
)

type ctxKey string

func PathParam(r *http.Request, key string) string {
	if params, ok := r.Context().Value(ctxKey("params")).(map[string]string); ok {
		return params[key]
	}
	return ""
}

func Route(mux *http.ServeMux, pattern string, handler http.Handler) {
	rePattern := "^" + regexp.MustCompile(`\{[^/]+\}`).ReplaceAllStringFunc(pattern, func(s string) string {
		return "([^/]+)"
	}) + "$"

	re := regexp.MustCompile(rePattern)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !re.MatchString(r.URL.Path) {
			return
		}

		paramNames := []string{}
		for _, part := range strings.Split(pattern, "/") {
			if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
				paramNames = append(paramNames, part[1:len(part)-1])
			}
		}

		matches := re.FindStringSubmatch(r.URL.Path)
		params := map[string]string{}
		for i, name := range paramNames {
			params[name] = matches[i+1]
		}

		ctx := context.WithValue(r.Context(), ctxKey("params"), params)
		handler.ServeHTTP(w, r.WithContext(ctx))
	})
}
