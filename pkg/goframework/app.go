package goframework

import (
	"encoding/json"
	"errors"
	"net/http"
	"path/filepath"
	"strings"
)

type App struct {
	router    *Router
	routesDir string
}

func New() *App {
	return &App{
		router: NewRouter(),
	}
}

func (a *App) LoadRoutes(root string) error {
	if root == "" {
		return errors.New("goframework: diretório de rotas vazio")
	}

	absRoot, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	regs := listRegistrations()
	if len(regs) == 0 {
		return errors.New("goframework: nenhuma rota registrada. certifique-se de importar os pacotes de rotas")
	}

	loaded := 0
	for _, reg := range regs {
		if !strings.HasPrefix(reg.file, absRoot) {
			continue
		}

		pattern, err := buildPatternFromFile(absRoot, reg.file)
		if err != nil {
			return err
		}

		err = a.router.AddRoute(reg.method, pattern, reg.handler)
		if err != nil {
			return err
		}
		loaded++
	}

	if loaded == 0 {
		return errors.New("goframework: nenhum handler encontrado dentro de " + absRoot)
	}

	a.routesDir = absRoot
	return nil
}

func (a *App) Listen(addr string) error {
	if a.router == nil {
		return errors.New("goframework: router não inicializado")
	}
	return http.ListenAndServe(addr, a)
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler, params, ok := a.router.Match(r.Method, r.URL.Path)
	if !ok {
		writeJSONError(w, http.StatusNotFound, "route_not_found")
		return
	}

	rw := &responseWriter{ResponseWriter: w}
	ctx := newContext(rw, r, params)
	err := handler(ctx)
	if err == nil {
		return
	}

	if rw.wrote {
		return
	}

	a.handleError(rw, err)
}

func (a *App) handleError(w http.ResponseWriter, err error) {
	switch typed := err.(type) {
	case *HTTPError:
		writeJSONError(w, typed.Status, typed.Message)
	default:
		writeJSONError(w, http.StatusInternalServerError, "internal_error")
	}
}

func writeJSONError(w http.ResponseWriter, status int, message string) {
	resp := map[string]string{"error": message}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(resp)
}

func buildPatternFromFile(root, file string) (string, error) {
	rel, err := filepath.Rel(root, file)
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(rel, "..") {
		return "", errors.New("goframework: rota fora do diretório raiz: " + file)
	}

	dir := filepath.Dir(rel)
	if dir == "." {
		dir = ""
	}

	segments := make([]string, 0)
	if dir != "" {
		for _, segment := range strings.Split(dir, string(filepath.Separator)) {
			if segment == "index" || segment == "" {
				continue
			}
			// Tratar pastas que começam com _ como parâmetros (:param)
			if strings.HasPrefix(segment, "_") {
				paramName := strings.TrimPrefix(segment, "_")
				segments = append(segments, ":"+paramName)
			} else {
				segments = append(segments, segment)
			}
		}
	}

	path := "/"
	if len(segments) > 0 {
		path = "/" + strings.Join(segments, "/")
	}
	return path, nil
}
