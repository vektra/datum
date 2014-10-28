package datum

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/ugorji/go/codec"
	"github.com/vektra/neko"
)

func TestMsgpackBackend(t *testing.T) {
	n := neko.Start(t)

	var mp *MsgpackBackend

	var ms MockBlobStore

	n.CheckMock(&ms.Mock)

	n.Setup(func() {
		mp = &MsgpackBackend{&ms}
	})

	n.It("stores new keys", func() {
		var data []byte

		doc := map[string]interface{}{"blah": "foo"}

		err := codec.NewEncoderBytes(&data, msgpackHandle).Encode(doc)
		require.NoError(t, err)

		ms.On("Get", "aabbcc", "default").Return([]byte(nil), nil)
		ms.On("Set", "aabbcc", "default", data).Return(nil)

		err = mp.Set("aabbcc", "default", "blah", "foo")
		require.NoError(t, err)
	})

	n.NIt("stores keys in an existing document", func() {
		var data []byte

		doc := map[string]interface{}{"name": "vektra"}

		err := codec.NewEncoderBytes(&data, msgpackHandle).Encode(doc)
		require.NoError(t, err)

		ms.On("Get", "aabbcc", "default").Return(data, nil)

		doc["blah"] = "foo"

		var data2 []byte

		err = codec.NewEncoderBytes(&data2, msgpackHandle).Encode(doc)
		require.NoError(t, err)

		ms.On("Set", "aabbcc", "default", data2).Return(nil)

		err = mp.Set("aabbcc", "default", "blah", "foo")
		require.NoError(t, err)
	})

	n.It("stores dotted keys as sub-maps", func() {
		var data []byte

		doc := map[string]interface{}{
			"sub": map[string]interface{}{
				"blah": "foo",
			},
		}

		err := codec.NewEncoderBytes(&data, msgpackHandle).Encode(doc)
		require.NoError(t, err)

		ms.On("Get", "aabbcc", "default").Return([]byte(nil), nil)
		ms.On("Set", "aabbcc", "default", data).Return(nil)

		err = mp.Set("aabbcc", "default", "sub.blah", "foo")
		require.NoError(t, err)
	})

	n.It("deletes keys when the value is nil", func() {
		var data []byte

		doc := map[string]interface{}{"blah": "foo"}

		err := codec.NewEncoderBytes(&data, msgpackHandle).Encode(doc)
		require.NoError(t, err)

		ms.On("Get", "aabbcc", "default").Return(data, nil)

		var data2 []byte

		delete(doc, "blah")

		err = codec.NewEncoderBytes(&data2, msgpackHandle).Encode(doc)
		require.NoError(t, err)

		ms.On("Set", "aabbcc", "default", data2).Return(nil)

		err = mp.Set("aabbcc", "default", "blah", nil)
		require.NoError(t, err)
	})

	n.It("deletes keys in a subtree when the value is nil", func() {
		var data []byte

		sub := map[string]interface{}{"bar": "foo", "qux": "kek"}

		doc := map[string]interface{}{"blah": sub}

		err := codec.NewEncoderBytes(&data, msgpackHandle).Encode(doc)
		require.NoError(t, err)

		ms.On("Get", "aabbcc", "default").Return(data, nil)

		var data2 []byte

		delete(sub, "bar")

		err = codec.NewEncoderBytes(&data2, msgpackHandle).Encode(doc)
		require.NoError(t, err)

		ms.On("Set", "aabbcc", "default", data2).Return(nil)

		err = mp.Set("aabbcc", "default", "blah.bar", nil)
		require.NoError(t, err)
	})

	n.It("deletes empty maps keys", func() {
		var data []byte

		sub := map[string]interface{}{"bar": "foo"}

		doc := map[string]interface{}{"blah": sub}

		err := codec.NewEncoderBytes(&data, msgpackHandle).Encode(doc)
		require.NoError(t, err)

		ms.On("Get", "aabbcc", "default").Return(data, nil)

		var data2 []byte

		delete(doc, "blah")

		err = codec.NewEncoderBytes(&data2, msgpackHandle).Encode(doc)
		require.NoError(t, err)

		ms.On("Set", "aabbcc", "default", data2).Return(nil)

		err = mp.Set("aabbcc", "default", "blah.bar", nil)
		require.NoError(t, err)
	})

	n.It("gets values", func() {
		var data []byte

		doc := map[string]interface{}{"blah": "foo"}

		err := codec.NewEncoderBytes(&data, msgpackHandle).Encode(doc)
		require.NoError(t, err)

		ms.On("Get", "aabbcc", "default").Return(data, nil)

		val, err := mp.Get("aabbcc", "default", "blah")
		require.NoError(t, err)

		assert.Equal(t, "foo", val)
	})

	n.It("gets dotted keys as sub-maps", func() {
		var data []byte

		doc := map[string]interface{}{
			"sub": map[string]interface{}{
				"blah": "foo",
			},
		}

		err := codec.NewEncoderBytes(&data, msgpackHandle).Encode(doc)
		require.NoError(t, err)

		ms.On("Get", "aabbcc", "default").Return(data, nil)

		val, err := mp.Get("aabbcc", "default", "sub.blah")
		require.NoError(t, err)

		assert.Equal(t, "foo", val)
	})

	n.It("returns a whole space given an empty key", func() {
		var data []byte

		doc := map[string]interface{}{
			"sub": map[string]interface{}{
				"blah": "foo",
			},
		}

		err := codec.NewEncoderBytes(&data, msgpackHandle).Encode(doc)
		require.NoError(t, err)

		ms.On("Get", "aabbcc", "default").Return(data, nil)

		val, err := mp.Get("aabbcc", "default", "")
		require.NoError(t, err)

		assert.Equal(t, doc, val)
	})

	n.It("stores an encrypted value", func() {
		var data []byte

		encVal := &EncryptedValue{Keyid: "a1b2c3", Value: []byte("foo")}

		doc := map[string]interface{}{"blah": encVal}

		err := codec.NewEncoderBytes(&data, msgpackHandle).Encode(doc)
		require.NoError(t, err)

		ms.On("Get", "aabbcc", "default").Return([]byte(nil), nil)
		ms.On("Set", "aabbcc", "default", data).Return(nil)

		err = mp.Set("aabbcc", "default", "blah", encVal)
		require.NoError(t, err)
	})

	n.It("retrieves an encrypted value", func() {
		var data []byte

		encVal := &EncryptedValue{Keyid: "a1b2c3", Value: []byte("foo")}

		doc := map[string]interface{}{"blah": encVal}

		err := codec.NewEncoderBytes(&data, msgpackHandle).Encode(doc)
		require.NoError(t, err)

		ms.On("Get", "aabbcc", "default").Return(data, nil)

		val, err := mp.Get("aabbcc", "default", "blah")
		require.NoError(t, err)

		assert.Equal(t, *encVal, val)
	})

	n.Meow()
}
