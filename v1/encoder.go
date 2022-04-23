package v1

import (
	"crypto/sha512"
	"encoding/base64"
)

const (
	LEN_BYTES          = 32
	CHECKSUM_LEN_BYTES = 4
)

func checkSum(data []byte) []byte {
	_checksums := sha512.Sum512_256(data)
	return _checksums[:]
}

func encode_address(data []byte) (addr string, e error) {
	if len(data) != LEN_BYTES {
		e = &LibraryError{"Address length is not 32"}
	}

	_addr := sha512.Sum512_256(data)
	checkSum := _addr[LEN_BYTES-CHECKSUM_LEN_BYTES:]
	_strAddr := base64.StdEncoding.EncodeToString(append(data, checkSum...))

	addr = _strAddr
	return
}
