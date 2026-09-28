// SPDX-License-Identifier: Unlicense

package tdata

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha512"
	"encoding/binary"
	"errors"
	"io"
)

const (
	tdesktopFileMagic    = "TDF$"
	tdesktopSessionMagic = 75
	tdesktopSaltSize     = 32
	sessionFileNameSize  = 16
	messageKeySize       = 16
	authKeySize          = 256
)

type TDataSession struct {
	AuthKey string `json:"authKey"`
	UserID  uint64 `json:"userId"`
	DC      uint32 `json:"dc"`
}

func GetServerAddress(dcID uint32) (string, error) {
	switch dcID {
	case 1:
		return "149.154.175.55", nil
	case 2:
		return "149.154.167.50", nil
	case 3:
		return "149.154.175.100", nil
	case 4:
		return "149.154.167.91", nil
	case 5:
		return "91.108.56.170", nil
	default:
		return "", errors.New("Invalid DC")
	}
}

func tdesktopFileHash(data []byte, versionBytes []byte) [md5.Size]byte {
	var length [4]byte
	binary.LittleEndian.PutUint32(length[:], uint32(len(data)))

	hashData := make([]byte, 0, len(data)+len(length)+len(versionBytes)+len(tdesktopFileMagic))
	hashData = append(hashData, data...)
	hashData = append(hashData, length[:]...)
	hashData = append(hashData, versionBytes...)
	hashData = append(hashData, tdesktopFileMagic...)

	return md5.Sum(hashData)
}

func calcOldAESKey(authKey []byte, msgKey []byte, client bool) ([]byte, []byte, error) {
	x := 8
	if client {
		x = 0
	}
	if len(authKey) < x+128 {
		return nil, nil, io.ErrUnexpectedEOF
	}
	if len(msgKey) != messageKeySize {
		return nil, nil, io.ErrUnexpectedEOF
	}

	sha1A := sha1Bytes(msgKey, authKey[x:x+32])
	sha1B := sha1Bytes(authKey[32+x:32+x+16], msgKey, authKey[48+x:48+x+16])
	sha1C := sha1Bytes(authKey[64+x:64+x+32], msgKey)
	sha1D := sha1Bytes(msgKey, authKey[96+x:96+x+32])

	aesKey := make([]byte, 0, 32)
	aesKey = append(aesKey, sha1A[:8]...)
	aesKey = append(aesKey, sha1B[8:20]...)
	aesKey = append(aesKey, sha1C[4:16]...)

	aesIV := make([]byte, 0, 32)
	aesIV = append(aesIV, sha1A[8:20]...)
	aesIV = append(aesIV, sha1B[:8]...)
	aesIV = append(aesIV, sha1C[16:20]...)
	aesIV = append(aesIV, sha1D[:8]...)

	return aesKey, aesIV, nil
}

func sha1Bytes(parts ...[]byte) []byte {
	h := sha1.New()
	for _, part := range parts {
		_, _ = h.Write(part)
	}
	return h.Sum(nil)
}

func sha512Bytes(parts ...[]byte) []byte {
	h := sha512.New()
	for _, part := range parts {
		_, _ = h.Write(part)
	}
	return h.Sum(nil)
}

func xorBlock(dst []byte, a []byte, b []byte) {
	for i := range dst {
		dst[i] = a[i] ^ b[i]
	}
}
