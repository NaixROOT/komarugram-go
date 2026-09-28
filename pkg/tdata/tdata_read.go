// SPDX-License-Identifier: Unlicense

package tdata

import (
	"bytes"
	"crypto/aes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

var tdesktopFileSuffixes = [...]string{"s", "0", "1"}

type TDataParent struct {
	tdesktopKey []byte
	sessions    []string
}

func isTdesktopKey(filename string) bool {
	if strings.HasPrefix(filename, "key_data") {
		for _, suffix := range tdesktopFileSuffixes {
			if strings.HasSuffix(filename, suffix) {
				return true
			}
		}
	}
	return false
}

func readTDesktopKey(inputKeyData []byte) ([]byte, error) {
	keyData, err := tdesktopOpen(inputKeyData)
	if err != nil {
		return nil, fmt.Errorf("newKeyData: %w", err)
	}

	salt, err := tdesktopReadBuffer(keyData)
	if err != nil {
		return nil, fmt.Errorf("tdesktopReadBuffer: %w", err)
	}
	if len(salt) != tdesktopSaltSize {
		return nil, errors.New("Length of salt is wrong!")
	}

	encryptedKey, err := tdesktopReadBuffer(keyData)
	if err != nil {
		return nil, fmt.Errorf("tdesktopReadBuffer: %w", err)
	}

	encryptedInfo, err := tdesktopReadBuffer(keyData)
	if err != nil {
		return nil, fmt.Errorf("tdesktopReadBuffer: %w", err)
	}

	passwordHash := sha512Bytes(salt, salt)
	passKey := pbkdf2.Key(passwordHash, salt, 1, authKeySize, sha512.New)

	keyReader, err := tdesktopDecrypt(encryptedKey, passKey)
	if err != nil {
		return nil, fmt.Errorf("tdesktopDecrypt: %w", err)
	}

	tdesktopKey, err := tdesktopReadBuffer(keyReader)
	if err != nil {
		return nil, fmt.Errorf("tdesktopReadBuffer: %w", err)
	}

	if err := readTDesktopInfo(encryptedInfo, tdesktopKey); err != nil {
		return nil, fmt.Errorf("readTDesktopInfo: %w", err)
	}

	return tdesktopKey, nil
}

func readTDesktopInfo(encryptedInfo []byte, tdesktopKey []byte) error {
	infoReader, err := tdesktopDecrypt(encryptedInfo, tdesktopKey)
	if err != nil {
		return err
	}

	info, err := tdesktopReadBuffer(infoReader)
	if err != nil {
		return err
	}

	_, err = newBinaryReader(info).readUint32BE()
	return err
}

type binaryReader struct {
	data   []byte
	offset int
}

func newBinaryReader(data []byte) *binaryReader {
	return &binaryReader{data: data}
}

func (r *binaryReader) read(length int) ([]byte, error) {
	remaining := r.remaining()
	if length < 0 {
		length = remaining
	}
	if length < 0 || length > remaining {
		return nil, io.ErrUnexpectedEOF
	}

	result := r.data[r.offset : r.offset+length]
	r.offset += length
	return result, nil
}

func (r *binaryReader) readUint32LE() (uint32, error) {
	data, err := r.read(4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(data), nil
}

func (r *binaryReader) readUint32BE() (uint32, error) {
	data, err := r.read(4)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(data), nil
}

func (r *binaryReader) buffer() []byte {
	return r.data
}

func (r *binaryReader) remaining() int {
	return len(r.data) - r.offset
}

func tdesktopReadBuffer(file *binaryReader) ([]byte, error) {
	lengthBytes, err := file.read(4)
	if err != nil {
		return nil, err
	}

	length := int32(binary.BigEndian.Uint32(lengthBytes))
	if length <= 0 {
		return []byte{}, nil
	}

	return file.read(min(file.remaining(), int(length)))
}

func tdesktopDecrypt(data []byte, authKey []byte) (*binaryReader, error) {
	if len(data) < messageKeySize {
		return nil, io.ErrUnexpectedEOF
	}

	messageKey := data[:messageKeySize]
	encryptedData := data[messageKeySize:]
	aesKey, aesIV, err := calcOldAESKey(authKey, messageKey, false)
	if err != nil {
		return nil, err
	}

	decryptedData, err := aesIGEDecrypt(encryptedData, aesKey, aesIV)
	if err != nil {
		return nil, err
	}

	decryptedHash := sha1.Sum(decryptedData)
	if subtle.ConstantTimeCompare(messageKey, decryptedHash[:messageKeySize]) != 1 {
		return nil, errors.New("msg_key mismatch")
	}

	return newBinaryReader(decryptedData), nil
}

func tdesktopOpenEncrypted(sessionData []byte, tdesktopKey []byte) (*binaryReader, error) {
	file, err := tdesktopOpen(sessionData)
	if err != nil {
		return nil, err
	}

	data, err := tdesktopReadBuffer(file)
	if err != nil {
		return nil, err
	}

	result, err := tdesktopDecrypt(data, tdesktopKey)
	if err != nil {
		return nil, err
	}

	length, err := result.readUint32LE()
	if err != nil {
		return nil, err
	}
	if length > uint32(len(result.buffer())) || length < 4 {
		return nil, errors.New("Wrong length")
	}

	return result, nil
}

func tdesktopOpen(input []byte) (*binaryReader, error) {
	reader := newBinaryReader(input)
	magic, err := reader.read(4)
	if err != nil {
		return nil, err
	}
	if string(magic) != tdesktopFileMagic {
		return nil, errors.New("Wrong MAGIC")
	}

	versionBytes, err := reader.read(4)
	if err != nil {
		return nil, err
	}

	data, err := reader.read(-1)
	if err != nil {
		return nil, err
	}
	if len(data) < md5.Size {
		return nil, errors.New("Wrong MD5")
	}

	expectedHash := data[len(data)-md5.Size:]
	data = data[:len(data)-md5.Size]

	actualHash := tdesktopFileHash(data, versionBytes)
	if !bytes.Equal(actualHash[:], expectedHash) {
		return nil, errors.New("Wrong MD5")
	}

	return newBinaryReader(data), nil
}

func readTDesktopSession(sessionData []byte, tdesktopKey []byte) (TDataSession, error) {
	main, err := tdesktopOpenEncrypted(sessionData, tdesktopKey)
	if err != nil {
		return TDataSession{}, err
	}

	magic, err := main.readUint32BE()
	if err != nil {
		return TDataSession{}, err
	}
	if magic != tdesktopSessionMagic {
		return TDataSession{}, errors.New("Unsupported magic version")
	}

	finalData, err := tdesktopReadBuffer(main)
	if err != nil {
		return TDataSession{}, err
	}

	final := newBinaryReader(finalData)
	if _, err := final.read(12); err != nil {
		return TDataSession{}, err
	}

	userID, err := final.readUint32BE()
	if err != nil {
		return TDataSession{}, err
	}

	mainDC, err := final.readUint32BE()
	if err != nil {
		return TDataSession{}, err
	}

	authKeyCount, err := final.readUint32BE()
	if err != nil {
		return TDataSession{}, err
	}

	for i := uint32(0); i < authKeyCount; i++ {
		dc, err := final.readUint32BE()
		if err != nil {
			return TDataSession{}, err
		}

		authKey, err := final.read(authKeySize)
		if err != nil {
			return TDataSession{}, err
		}

		if dc == mainDC {
			return TDataSession{
				AuthKey: hex.EncodeToString(authKey),
				DC:      dc,
				UserID:  uint64(userID),
			}, nil
		}
	}

	return TDataSession{}, io.EOF
}

func aesIGEDecrypt(ciphertext []byte, key []byte, iv []byte) ([]byte, error) {
	if len(iv) != aes.BlockSize*2 {
		return nil, io.ErrUnexpectedEOF
	}
	if len(ciphertext)%aes.BlockSize != 0 {
		return nil, io.ErrUnexpectedEOF
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	plainText := make([]byte, len(ciphertext))
	prevCipher := append([]byte(nil), iv[:aes.BlockSize]...)
	prevPlain := append([]byte(nil), iv[aes.BlockSize:]...)
	tmp := make([]byte, aes.BlockSize)
	plainBlock := make([]byte, aes.BlockSize)

	for offset := 0; offset < len(ciphertext); offset += aes.BlockSize {
		cipherBlock := ciphertext[offset : offset+aes.BlockSize]

		xorBlock(tmp, cipherBlock, prevPlain)
		block.Decrypt(tmp, tmp)
		xorBlock(plainBlock, tmp, prevCipher)

		copy(plainText[offset:offset+aes.BlockSize], plainBlock)
		copy(prevPlain, plainBlock)
		copy(prevCipher, cipherBlock)
	}

	return plainText, nil
}
