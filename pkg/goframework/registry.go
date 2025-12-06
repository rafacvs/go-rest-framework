package goframework

import (
	"errors"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type registration struct {
	file    string
	method  string
	handler HandlerFunc
}

var (
	registryMu sync.Mutex
	routesReg  = make(map[string]registration)
)

func Register(handler HandlerFunc) {
	if handler == nil {
		panic("goframework: handler nil")
	}

	_, file, _, ok := runtime.Caller(1)
	if !ok {
		panic("goframework: não foi possível determinar o arquivo do handler")
	}

	absFile, err := filepath.Abs(file)
	if err != nil {
		panic("goframework: falha ao resolver caminho do handler: " + err.Error())
	}

	method, err := methodFromFilename(absFile)
	if err != nil {
		panic(err.Error())
	}

	registryMu.Lock()
	defer registryMu.Unlock()

	routesReg[absFile] = registration{
		file:    absFile,
		method:  method,
		handler: handler,
	}
}

func listRegistrations() []registration {
	registryMu.Lock()
	defer registryMu.Unlock()

	list := make([]registration, 0, len(routesReg))
	for _, reg := range routesReg {
		list = append(list, reg)
	}
	return list
}

func methodFromFilename(path string) (string, error) {
	base := filepath.Base(path)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	switch strings.ToLower(name) {
	case "get":
		return "GET", nil
	case "post":
		return "POST", nil
	case "put":
		return "PUT", nil
	case "patch":
		return "PATCH", nil
	case "delete":
		return "DELETE", nil
	case "options":
		return "OPTIONS", nil
	default:
		return "", errors.New("goframework: nome de arquivo não reconhecido para método HTTP: " + name)
	}
}
