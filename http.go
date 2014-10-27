package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/bmizerany/pat"
	"github.com/pelletier/go-toml"
)
import "net/http"

type TokenGenerator interface {
	NewToken() string
}

type Backend interface {
	Set(token string, space string, key string, val interface{}) error
	Get(token string, space string, key string) (interface{}, error)
}

type HTTPApi struct {
	tg TokenGenerator
	be Backend

	mux *pat.PatternServeMux
}

func NewHTTPApi(tg TokenGenerator, be Backend) *HTTPApi {
	h := &HTTPApi{tg, be, pat.New()}

	h.mux.Post("/create", http.HandlerFunc(h.create))

	h.mux.Put("/:token/~:space", http.HandlerFunc(h.put3))
	h.mux.Put("/:token/~:space/", http.HandlerFunc(h.put3))
	h.mux.Put("/~:space/", http.HandlerFunc(h.put4))
	h.mux.Put("/:key", http.HandlerFunc(h.put1))
	h.mux.Put("/:token/", http.HandlerFunc(h.put2))

	h.mux.Del("/:token/~:space", http.HandlerFunc(h.del3))
	h.mux.Del("/:token/~:space/", http.HandlerFunc(h.del3))
	h.mux.Del("/~:space/", http.HandlerFunc(h.del4))
	h.mux.Del("/:key", http.HandlerFunc(h.del1))
	h.mux.Del("/:token/", http.HandlerFunc(h.del2))

	h.mux.Get("/:token/~:space", http.HandlerFunc(h.get2))
	h.mux.Get("/:token/~:space/", http.HandlerFunc(h.get2))
	h.mux.Get("/~:space", http.HandlerFunc(h.get3))
	h.mux.Get("/~:space/", http.HandlerFunc(h.get3))
	h.mux.Get("/:token", http.HandlerFunc(h.get1))
	h.mux.Get("/:token/", http.HandlerFunc(h.get1))

	return h
}

func (h *HTTPApi) create(w http.ResponseWriter, req *http.Request) {
	token := h.tg.NewToken()

	fmt.Fprintf(w, "%s\n", token)
}

func (h *HTTPApi) put1(w http.ResponseWriter, req *http.Request) {
	var (
		headerToken = req.Header.Get("Config-Token")
		key         = req.URL.Query().Get(":key")
	)

	h.put(headerToken, "default", key, w, req)
}

func (h *HTTPApi) put2(w http.ResponseWriter, req *http.Request) {
	var (
		headerToken = req.Header.Get("Config-Token")
		token       = req.URL.Query().Get(":token")
	)

	var key string

	if headerToken == "" {
		key = pat.Tail("/:token/", req.URL.Path)
	} else {
		token = headerToken
		key = req.URL.Path[1:]
	}

	h.put(token, "default", key, w, req)
}

func (h *HTTPApi) put3(w http.ResponseWriter, req *http.Request) {
	var (
		token = req.URL.Query().Get(":token")
		space = req.URL.Query().Get(":space")
	)

	key := pat.Tail("/:token/~:space/", req.URL.Path)

	h.put(token, space, key, w, req)
}

func (h *HTTPApi) put4(w http.ResponseWriter, req *http.Request) {
	var (
		headerToken = req.Header.Get("Config-Token")
		space       = req.URL.Query().Get(":space")
		key         = pat.Tail("/~:space/", req.URL.Path)
	)

	h.put(headerToken, space, key, w, req)
}

func (h *HTTPApi) put(token, space, key string, w http.ResponseWriter, req *http.Request) {
	var (
		val interface{}
		err error
	)

	var asJson bool

	ext := filepath.Ext(key)
	switch ext {
	case "json":
		asJson = true
	}

	asJson = req.Header.Get("Content-Type") == "application/json"

	if ext != "" {
		key = key[:len(key)-len(ext)-1]
	}

	key = strings.Replace(key, "/", ".", -1)

	if asJson {
		err = json.NewDecoder(req.Body).Decode(&val)
	} else {
		var bytes []byte

		bytes, err = ioutil.ReadAll(req.Body)
		if err == nil {
			val = string(bytes)
		}
	}

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	err = h.be.Set(token, space, key, val)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func (h *HTTPApi) del1(w http.ResponseWriter, req *http.Request) {
	var (
		headerToken = req.Header.Get("Config-Token")
		key         = req.URL.Query().Get(":key")
	)

	h.del(headerToken, "default", key, w, req)
}

func (h *HTTPApi) del2(w http.ResponseWriter, req *http.Request) {
	var (
		headerToken = req.Header.Get("Config-Token")
		token       = req.URL.Query().Get(":token")
	)

	var key string

	if headerToken == "" {
		key = pat.Tail("/:token/", req.URL.Path)
	} else {
		token = headerToken
		key = req.URL.Path[1:]
	}

	h.del(token, "default", key, w, req)
}

func (h *HTTPApi) del3(w http.ResponseWriter, req *http.Request) {
	var (
		token = req.URL.Query().Get(":token")
		space = req.URL.Query().Get(":space")
	)

	key := pat.Tail("/:token/~:space/", req.URL.Path)

	h.del(token, space, key, w, req)
}

func (h *HTTPApi) del4(w http.ResponseWriter, req *http.Request) {
	var (
		headerToken = req.Header.Get("Config-Token")
		space       = req.URL.Query().Get(":space")
		key         = pat.Tail("/~:space/", req.URL.Path)
	)

	h.del(headerToken, space, key, w, req)
}

func (h *HTTPApi) del(token, space, key string, w http.ResponseWriter, req *http.Request) {
	key = strings.Replace(key, "/", ".", -1)

	err := h.be.Set(token, space, key, nil)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}
}

func (h *HTTPApi) get2(w http.ResponseWriter, req *http.Request) {
	var (
		token = req.URL.Query().Get(":token")
		space = req.URL.Query().Get(":space")
	)

	key := pat.Tail("/:token/~:space/", req.URL.Path)

	h.get(token, space, key, w, req)
}

func (h *HTTPApi) get1(w http.ResponseWriter, req *http.Request) {
	var (
		headerToken = req.Header.Get("Config-Token")
		token       = req.URL.Query().Get(":token")
	)

	var key string

	if headerToken != "" {
		key = req.URL.Path[1:]
		token = headerToken
	} else {
		key = pat.Tail("/:token/", req.URL.Path)
	}

	h.get(token, "default", key, w, req)
}

func (h *HTTPApi) get3(w http.ResponseWriter, req *http.Request) {
	var (
		headerToken = req.Header.Get("Config-Token")
		space       = req.URL.Query().Get(":space")
	)

	key := pat.Tail("/~:space/", req.URL.Path)

	h.get(headerToken, space, key, w, req)
}

func (h *HTTPApi) get0(w http.ResponseWriter, req *http.Request) {
	fmt.Printf("get0!\n")

	h.get("", "default", req.URL.Path, w, req)
}

func (h *HTTPApi) get(token, space, key string, w http.ResponseWriter, req *http.Request) {
	headerToken := req.Header.Get("Config-Token")

	if token == "" {
		token = headerToken
	}

	if token == "" {
		http.Error(w, "no token provided", 400)
		return
	}

	if space == "" {
		space = "default"
	}

	var asToml bool

	asJson := req.Header.Get("Accept") == "application/json"

	var ext string

	if key == "" {
		ext = filepath.Ext(space)
		space = space[:len(space)-len(ext)]
	} else {
		ext = filepath.Ext(key)
		key = key[:len(key)-len(ext)]
	}

	switch ext {
	case ".json":
		asJson = true
	case ".toml":
		asToml = true
	}

	key = strings.Replace(key, "/", ".", -1)

	val, err := h.be.Get(token, space, key)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	if mapVal, ok := val.(map[string]interface{}); ok {
		if asToml {
			tree := toml.TreeFromMap(mapVal)
			fmt.Fprintf(w, "%s", tree.ToString())
		} else {
			json.NewEncoder(w).Encode(mapVal)
		}
	} else {
		if asJson {
			json.NewEncoder(w).Encode(val)
		} else {
			fmt.Fprintf(w, "%s\n", val)
		}
	}
}

func (h *HTTPApi) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	h.mux.ServeHTTP(w, req)
}
