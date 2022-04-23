package v1

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"

	"github.com/algorand/go-algorand-sdk/client/v2/common/models"
)

func ExtractInt(state map[string]models.TealValue, key string) uint64 {
	elem, ok := state[key]
	if ok {
		return elem.Uint
	}

	elem, ok = state[encodeKey(key)]
	if ok {
		return elem.Uint
	}

	return 0
}

func encodeKey(key string) string {
	_bytes := []byte(key)
	return base64.StdEncoding.EncodeToString(_bytes)
}

func TealValueArrayToMap(data []models.TealKeyValue) map[string]models.TealValue {
	var mapData map[string]models.TealValue
	for _, v := range data {
		mapData[v.Key] = v.Value
	}

	return mapData
}

func intToBytes(v uint64) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, v)
	return buf.Bytes()
}

func IntToStateKey(value int64) string {
	_paddingBytes := []byte("o")
	_valueBytes := intToBytes(uint64(value))
	_bytes := append(_paddingBytes, _valueBytes...)

	return base64.StdEncoding.EncodeToString(_bytes)
}
