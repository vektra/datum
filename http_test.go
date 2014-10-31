package datum

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vektra/neko"
)

func TestHTTP(t *testing.T) {
	n := neko.Start(t)

	var h *HTTPApi

	var (
		tg MockTokenGenerator
		be MockBackend
	)

	n.CheckMock(&tg.Mock)
	n.CheckMock(&be.Mock)

	n.Setup(func() {
		h = NewHTTPApi(&tg, &be)
	})

	n.It("can create a new token", func() {
		req, err := http.NewRequest("POST", "/create", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()

		token := "aabbcc"

		tg.On("NewToken").Return(token)

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, token+"\n", w.Body.String())
	})

	n.It("can create a one-use token for another view", func() {
		parent := "aabbcc"

		req, err := http.NewRequest("POST", "/create/onetime/aabbcc", nil)
		require.NoError(t, err)

		w := httptest.NewRecorder()

		token := "ddeeff"

		tg.On("NewToken").Return(token)

		be.On("Set", "_", "onetime", token, parent).Return(nil)

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, token+"\n", w.Body.String())
	})

	n.It("can add a key to a doc", func() {
		req, err := http.NewRequest("PUT", "/aabbcc/~def/blah", strings.NewReader("foo"))
		require.NoError(t, err)

		be.On("Set", "aabbcc", "def", "blah", "foo").Return(nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	n.It("can add a subtree key to a doc", func() {
		req, err := http.NewRequest("PUT", "/aabbcc/~def/blah/bar", strings.NewReader("foo"))
		require.NoError(t, err)

		be.On("Set", "aabbcc", "def", "blah.bar", "foo").Return(nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	n.It("can add a key to a as json", func() {
		req, err := http.NewRequest("PUT", "/aabbcc/~def/blah", strings.NewReader("1"))
		require.NoError(t, err)

		req.Header.Add("Content-Type", "application/json")

		be.On("Set", "aabbcc", "def", "blah", 1).Return(nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	n.It("can retrieve a key", func() {
		req, err := http.NewRequest("GET", "/aabbcc/~def/blah", nil)
		require.NoError(t, err)

		be.On("Get", "aabbcc", "def", "blah").Return([]byte("foo"), nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		assert.Equal(t, "foo\n", w.Body.String())
	})

	n.It("can returns a tree in json by default", func() {
		req, err := http.NewRequest("GET", "/aabbcc/~def/blah", nil)
		require.NoError(t, err)

		doc := map[string]interface{}{
			"bar": "foo",
		}

		be.On("Get", "aabbcc", "def", "blah").Return(doc, nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		assert.Equal(t, `{"bar":"foo"}`+"\n", w.Body.String())
	})

	n.It("can return a value in a subtree", func() {
		req, err := http.NewRequest("GET", "/aabbcc/~def/blah/bar", nil)
		require.NoError(t, err)

		be.On("Get", "aabbcc", "def", "blah.bar").Return("foo", nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		assert.Equal(t, "foo\n", w.Body.String())
	})

	n.It("can retrieve a key as JSON if requested", func() {
		req, err := http.NewRequest("GET", "/aabbcc/~def/blah", nil)
		require.NoError(t, err)

		req.Header.Add("Accept", "application/json")

		be.On("Get", "aabbcc", "def", "blah").Return("foo", nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		assert.Equal(t, `"foo"`+"\n", w.Body.String())
	})

	n.It("can retrieve a key as JSON if requested", func() {
		req, err := http.NewRequest("GET", "/aabbcc/~def/blah.json", nil)
		require.NoError(t, err)

		be.On("Get", "aabbcc", "def", "blah").Return("foo", nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		assert.Equal(t, `"foo"`+"\n", w.Body.String())
	})

	n.It("can returns a tree in TOML if requested", func() {
		req, err := http.NewRequest("GET", "/aabbcc/~def/blah.toml", nil)
		require.NoError(t, err)

		doc := map[string]interface{}{
			"bar": "foo",
		}

		be.On("Get", "aabbcc", "def", "blah").Return(doc, nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		expected := "bar = \"foo\"\n"

		assert.Equal(t, expected, w.Body.String())
	})

	n.It("can returns a sub tree in TOML if requested", func() {
		req, err := http.NewRequest("GET", "/aabbcc/~def/blah.toml", nil)
		require.NoError(t, err)

		doc := map[string]interface{}{
			"bar": map[string]interface{}{
				"qux": "foo",
			},
		}

		be.On("Get", "aabbcc", "def", "blah").Return(doc, nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		expected := "\n[bar]\n  qux = \"foo\"\n"

		assert.Equal(t, expected, w.Body.String())
	})

	n.It("sets a default space if none specified", func() {
		req, err := http.NewRequest("PUT", "/aabbcc/blah", strings.NewReader("foo"))
		require.NoError(t, err)

		be.On("Set", "aabbcc", "default", "blah", "foo").Return(nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	n.It("sets a default space if none specified with header token", func() {
		req, err := http.NewRequest("PUT", "/blah", strings.NewReader("foo"))
		require.NoError(t, err)

		req.Header.Set("Config-Token", "aabbcc")

		be.On("Set", "aabbcc", "default", "blah", "foo").Return(nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	n.It("sets a subtree key in the default space with header token", func() {
		req, err := http.NewRequest("PUT", "/blah/bar", strings.NewReader("foo"))
		require.NoError(t, err)

		req.Header.Set("Config-Token", "aabbcc")

		be.On("Set", "aabbcc", "default", "blah.bar", "foo").Return(nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	n.It("sets a named space if none specified with header token", func() {
		req, err := http.NewRequest("PUT", "/~here/blah", strings.NewReader("foo"))
		require.NoError(t, err)

		req.Header.Set("Config-Token", "aabbcc")

		be.On("Set", "aabbcc", "here", "blah", "foo").Return(nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	n.It("sets a subtree key in a named space with header token", func() {
		req, err := http.NewRequest("PUT", "/~here/blah/bar", strings.NewReader("foo"))
		require.NoError(t, err)

		req.Header.Set("Config-Token", "aabbcc")

		be.On("Set", "aabbcc", "here", "blah.bar", "foo").Return(nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	n.It("can retrieve a key from the default space", func() {
		req, err := http.NewRequest("GET", "/aabbcc/blah", nil)
		require.NoError(t, err)

		be.On("Get", "aabbcc", "default", "blah").Return([]byte("foo"), nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		assert.Equal(t, "foo\n", w.Body.String())
	})

	n.It("can retrieve a key from the default space using a header token", func() {
		req, err := http.NewRequest("GET", "/blah", nil)
		require.NoError(t, err)

		req.Header.Set("Config-Token", "aabbcc")

		be.On("Get", "aabbcc", "default", "blah").Return([]byte("foo"), nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		assert.Equal(t, "foo\n", w.Body.String())
	})

	n.It("can return a value in a subtree", func() {
		req, err := http.NewRequest("GET", "/blah/bar", nil)
		require.NoError(t, err)

		req.Header.Set("Config-Token", "aabbcc")

		be.On("Get", "aabbcc", "default", "blah.bar").Return("foo", nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		assert.Equal(t, "foo\n", w.Body.String())
	})

	n.It("can return a value in a subtree with a listed token", func() {
		req, err := http.NewRequest("GET", "/aabbcc/blah/bar", nil)
		require.NoError(t, err)

		be.On("Get", "aabbcc", "default", "blah.bar").Return("foo", nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		assert.Equal(t, "foo\n", w.Body.String())
	})

	n.It("can retrieve a key from the name space using a header token", func() {
		req, err := http.NewRequest("GET", "/~current/blah", nil)
		require.NoError(t, err)

		req.Header.Set("Config-Token", "aabbcc")

		be.On("Get", "aabbcc", "current", "blah").Return([]byte("foo"), nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		assert.Equal(t, "foo\n", w.Body.String())
	})

	n.It("can retrieve a subtree key from the name space using a header token", func() {
		req, err := http.NewRequest("GET", "/~current/blah/bar", nil)
		require.NoError(t, err)

		req.Header.Set("Config-Token", "aabbcc")

		be.On("Get", "aabbcc", "current", "blah.bar").Return([]byte("foo"), nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		assert.Equal(t, "foo\n", w.Body.String())
	})

	n.It("can retrieve a whole named space", func() {
		req, err := http.NewRequest("GET", "/~current", nil)
		require.NoError(t, err)

		req.Header.Set("Config-Token", "aabbcc")

		doc := map[string]interface{}{
			"blah": "foo",
		}

		be.On("Get", "aabbcc", "current", "").Return(doc, nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		assert.Equal(t, `{"blah":"foo"}`+"\n", w.Body.String())
	})

	n.It("can delete a key to a doc", func() {
		req, err := http.NewRequest("DELETE", "/aabbcc/~def/blah", nil)
		require.NoError(t, err)

		be.On("Set", "aabbcc", "def", "blah", nil).Return(nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	n.It("can delete a key in a subtree", func() {
		req, err := http.NewRequest("DELETE", "/aabbcc/~def/blah/bar", nil)
		require.NoError(t, err)

		be.On("Set", "aabbcc", "def", "blah.bar", nil).Return(nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	n.It("can delete a key in the default space", func() {
		req, err := http.NewRequest("DELETE", "/aabbcc/blah", nil)
		require.NoError(t, err)

		be.On("Set", "aabbcc", "default", "blah", nil).Return(nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	n.It("can delete a key in a subtree in the default space", func() {
		req, err := http.NewRequest("DELETE", "/aabbcc/blah/bar", nil)
		require.NoError(t, err)

		be.On("Set", "aabbcc", "default", "blah.bar", nil).Return(nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	n.It("can delete a key to a doc with a header token", func() {
		req, err := http.NewRequest("DELETE", "/~def/blah", nil)
		require.NoError(t, err)

		req.Header.Set("Config-Token", "aabbcc")

		be.On("Set", "aabbcc", "def", "blah", nil).Return(nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	n.It("can delete a subtree key to a doc with a header token", func() {
		req, err := http.NewRequest("DELETE", "/~def/blah/bar", nil)
		require.NoError(t, err)

		req.Header.Set("Config-Token", "aabbcc")

		be.On("Set", "aabbcc", "def", "blah.bar", nil).Return(nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	n.It("can delete a key to a doc with a header token in default", func() {
		req, err := http.NewRequest("DELETE", "/blah", nil)
		require.NoError(t, err)

		req.Header.Set("Config-Token", "aabbcc")

		be.On("Set", "aabbcc", "default", "blah", nil).Return(nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	n.It("can delete a subtree key to a doc with a header token in default", func() {
		req, err := http.NewRequest("DELETE", "/blah/bar", nil)
		require.NoError(t, err)

		req.Header.Set("Config-Token", "aabbcc")

		be.On("Set", "aabbcc", "default", "blah.bar", nil).Return(nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	n.It("can returns a space in TOML if requested", func() {
		req, err := http.NewRequest("GET", "/aabbcc/~def.toml", nil)
		require.NoError(t, err)

		doc := map[string]interface{}{
			"bar": "foo",
		}

		be.On("Get", "aabbcc", "def", "").Return(doc, nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		expected := "bar = \"foo\"\n"

		assert.Equal(t, expected, w.Body.String())
	})

	n.It("maps view tokens to their parent on get", func() {
		req, err := http.NewRequest("GET", "/v-ddeeff/~def/bar", nil)
		require.NoError(t, err)

		be.On("Get", "_", "views", "v-ddeeff").Return("aabbcc", nil)

		be.On("Get", "aabbcc", "def", "bar").Return("foo", nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		assert.Equal(t, "foo\n", w.Body.String())
	})

	n.It("maps view tokens to their parent on set", func() {
		req, err := http.NewRequest("PUT", "/v-ddeeff/~def/bar", strings.NewReader("foo"))
		require.NoError(t, err)

		be.On("Get", "_", "views", "v-ddeeff").Return("aabbcc", nil)

		be.On("Set", "aabbcc", "def", "bar", "foo").Return(nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	n.It("deletes one-time tokens on first get", func() {
		req, err := http.NewRequest("GET", "/o-ddeeff/~def/bar", nil)
		require.NoError(t, err)

		be.On("Get", "_", "onetime", "o-ddeeff").Return("aabbcc", nil)
		be.On("Set", "_", "onetime", "o-ddeeff", nil).Return(nil)

		be.On("Get", "aabbcc", "def", "bar").Return("foo", nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)

		assert.Equal(t, "foo\n", w.Body.String())
	})

	n.It("deletes one-time tokens on first set", func() {
		req, err := http.NewRequest("PUT", "/o-ddeeff/~def/bar", strings.NewReader("foo"))
		require.NoError(t, err)

		be.On("Get", "_", "onetime", "o-ddeeff").Return("aabbcc", nil)
		be.On("Set", "_", "onetime", "o-ddeeff", nil).Return(nil)

		be.On("Set", "aabbcc", "def", "bar", "foo").Return(nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	n.It("records an encrypted value along with the keyid", func() {
		req, err := http.NewRequest("PUT", "/aabbcc/~def/bar", strings.NewReader("foo"))
		require.NoError(t, err)

		req.Header.Set("Config-Encryption-KeyID", "a1b2c3")

		encVal := &EncryptedValue{
			Value: []byte("foo"),
			Keyid: "a1b2c3",
		}

		be.On("Set", "aabbcc", "def", "bar", encVal).Return(nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
	})

	n.It("returns an encrypted value along with the keyid", func() {
		req, err := http.NewRequest("GET", "/aabbcc/~def/bar", nil)
		require.NoError(t, err)

		encVal := EncryptedValue{
			Value: []byte("foo"),
			Keyid: "a1b2c3",
		}

		be.On("Get", "aabbcc", "def", "bar").Return(encVal, nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "foo", w.Body.String())
		assert.Equal(t, "a1b2c3", w.Header().Get("Config-Encryption-KeyID"))
	})

	n.It("returns an encrypted value as json", func() {
		req, err := http.NewRequest("GET", "/aabbcc/~def/bar", nil)
		require.NoError(t, err)

		req.Header.Set("Accept", "application/json")

		encVal := EncryptedValue{
			Value: []byte("foo"),
			Keyid: "a1b2c3",
		}

		be.On("Get", "aabbcc", "def", "bar").Return(encVal, nil)

		w := httptest.NewRecorder()

		h.ServeHTTP(w, req)

		json := fmt.Sprintf(`{"keyid":"%s","value":"%s"}%s`,
			"a1b2c3", base64.StdEncoding.EncodeToString(encVal.Value), "\n")

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, json, w.Body.String())
	})

	n.Meow()
}
