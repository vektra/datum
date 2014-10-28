package datum

import (
	"bufio"
	"bytes"
	"fmt"
	"reflect"
	"strings"

	"github.com/ugorji/go/codec"
)

type BlobStore interface {
	Set(key string, space string, val []byte) error
	Get(key string, space string) ([]byte, error)
}

type MsgpackBackend struct {
	store BlobStore
}

func NewMsgpackBackend(store BlobStore) *MsgpackBackend {
	return &MsgpackBackend{store}
}

var msgpackHandle = &codec.MsgpackHandle{}

func init() {
	msgpackHandle.RawToString = true
	msgpackHandle.MapType = reflect.TypeOf(map[string]interface{}{})
	msgpackHandle.WriteExt = true

	msgpackHandle.SetExt(reflect.TypeOf(EncryptedValue{}), 0x47, &encryptedValExt{})
}

type encryptedValExt struct{}

func (_ *encryptedValExt) WriteExt(v reflect.Value) []byte {
	encVal := v.Interface().(EncryptedValue)

	var buf bytes.Buffer

	buf.WriteString(encVal.Keyid)
	buf.WriteString("\n")
	buf.Write(encVal.Value)

	return buf.Bytes()
}

func (_ *encryptedValExt) ReadExt(v reflect.Value, b []byte) {
	encVal := v.Interface().(EncryptedValue)

	buffer := bufio.NewReader(bytes.NewReader(b))

	keyid, _ := buffer.ReadString('\n')

	val := make([]byte, len(b)-len(keyid))

	buffer.Read(val)

	encVal.Keyid = keyid[0 : len(keyid)-1]
	encVal.Value = val

	v.Set(reflect.ValueOf(encVal))
}

func (_ *encryptedValExt) ConvertExt(v reflect.Value) interface{} {
	return v.Interface()
}

func (_ *encryptedValExt) UpdateExt(v reflect.Value, i interface{}) {}

func (m *MsgpackBackend) findSub(
	doc map[string]interface{},
	keys []string,
) (map[string]interface{}, error) {

	pos := doc

	for _, k := range keys {
		if val, ok := pos[k]; ok {
			if m, ok := val.(map[string]interface{}); ok {
				pos = m
			} else {
				return nil, fmt.Errorf("%s is not a map", k)
			}
		} else {
			m := make(map[string]interface{})
			pos[k] = m
			pos = m
		}
	}

	return pos, nil
}

func (m *MsgpackBackend) prune(pos map[string]interface{}) {
	var toRemove []string

	for k, v := range pos {
		if sub, ok := v.(map[string]interface{}); ok {
			if len(sub) == 0 {
				toRemove = append(toRemove, k)
			} else {
				m.prune(sub)
			}
		}
	}

	for _, k := range toRemove {
		delete(pos, k)
	}
}

func (m *MsgpackBackend) Set(token, space, key string, val interface{}) error {
	blob, err := m.store.Get(token, space)
	if err != nil {
		return err
	}

	var doc map[string]interface{}

	if blob == nil {
		doc = make(map[string]interface{})
	} else {
		err = codec.NewDecoderBytes(blob, msgpackHandle).Decode(&doc)
		if err != nil {
			return err
		}
	}

	parts := strings.Split(key, ".")

	key = parts[len(parts)-1]

	var pos map[string]interface{}

	if len(parts) == 1 {
		pos = doc
	} else {
		pos, err = m.findSub(doc, parts[:len(parts)-1])
		if err != nil {
			return err
		}
	}

	if val == nil {
		delete(pos, key)
		m.prune(doc)
	} else {
		pos[key] = val
	}

	var data []byte

	err = codec.NewEncoderBytes(&data, msgpackHandle).Encode(doc)
	if err != nil {
		return err
	}

	return m.store.Set(token, space, data)
}

func (m *MsgpackBackend) Get(token, space, key string) (interface{}, error) {
	blob, err := m.store.Get(token, space)
	if err != nil {
		return nil, err
	}

	if blob == nil {
		return nil, nil
	}

	var doc map[string]interface{}

	err = codec.NewDecoderBytes(blob, msgpackHandle).Decode(&doc)
	if err != nil {
		return nil, err
	}

	if key == "" {
		return doc, nil
	}

	parts := strings.Split(key, ".")

	key = parts[len(parts)-1]

	var pos map[string]interface{}

	if len(parts) == 1 {
		pos = doc
	} else {
		pos, err = m.findSub(doc, parts[:len(parts)-1])
		if err != nil {
			return nil, err
		}
	}

	return pos[key], nil
}
