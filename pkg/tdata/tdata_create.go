// SPDX-License-Identifier: Unlicense

package tdata

import (
	"archive/zip"
	"crypto/aes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha512"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	tdesktopFileVersion         = 1
	tdesktopDataName            = "data"
	tdesktopMapFileName         = "map"
	tdesktopKeyDataName         = "key_data"
	tdesktopBufferHeaderSize    = 4
	tdesktopSingleAccountCount  = 1
	tdesktopActiveAccountIndex  = 0
	tdesktopWideIDMarker        = uint32(0xffffffff)
	tdesktopNoDestroyedKeyCount = 0
	tdesktopMaxAccountCount     = 3
	tdesktopZipRootName         = "tdata"
	tdesktopDirMode             = fs.FileMode(0o700)
	tdesktopFileMode            = fs.FileMode(0o600)
)

type tdesktopSessionFiles struct {
	keyData  []byte
	accounts []tdesktopAccountFiles
}

type tdesktopAccountFiles struct {
	dataKey     string
	sessionData []byte
	mapData     []byte
}

func buildTDesktopSessionFiles(sessions []TDataSession) (tdesktopSessionFiles, error) {
	authKeys, err := decodeTDataSessionAuthKeys(sessions)
	if err != nil {
		return tdesktopSessionFiles{}, err
	}

	salt, err := randomBytes(tdesktopSaltSize)
	if err != nil {
		return tdesktopSessionFiles{}, err
	}

	tdesktopKey, err := randomBytes(authKeySize)
	if err != nil {
		return tdesktopSessionFiles{}, err
	}

	passKey := tdesktopPassKey(salt)
	keyData, err := buildTDesktopKeyData(salt, passKey, tdesktopKey, len(sessions))
	if err != nil {
		return tdesktopSessionFiles{}, err
	}

	accounts := make([]tdesktopAccountFiles, 0, len(sessions))
	for i, session := range sessions {
		account, err := buildTDesktopAccountFiles(i, session, authKeys[i], tdesktopKey)
		if err != nil {
			return tdesktopSessionFiles{}, err
		}
		accounts = append(accounts, account)
	}

	return tdesktopSessionFiles{
		keyData:  keyData,
		accounts: accounts,
	}, nil
}

func decodeTDataSessionAuthKeys(sessions []TDataSession) ([][]byte, error) {
	if len(sessions) == 0 {
		return nil, errors.New("at least one session is required")
	}
	if len(sessions) > tdesktopMaxAccountCount {
		return nil, fmt.Errorf("tdata supports up to %d sessions", tdesktopMaxAccountCount)
	}

	authKeys := make([][]byte, 0, len(sessions))
	for i, session := range sessions {
		if err := validateTDesktopDC(session.DC); err != nil {
			return nil, fmt.Errorf("session %d: %w", i, err)
		}

		authKey, err := decodeAuthKey(session.AuthKey)
		if err != nil {
			return nil, fmt.Errorf("session %d: %w", i, err)
		}
		authKeys = append(authKeys, authKey)
	}
	return authKeys, nil
}

func buildTDesktopAccountFiles(
	accountIndex int,
	session TDataSession,
	authKey []byte,
	tdesktopKey []byte,
) (tdesktopAccountFiles, error) {
	sessionData, err := buildTDesktopSessionData(session, authKey, tdesktopKey)
	if err != nil {
		return tdesktopAccountFiles{}, err
	}

	mapData, err := buildTDesktopEmptyMapData(tdesktopKey)
	if err != nil {
		return tdesktopAccountFiles{}, err
	}

	return tdesktopAccountFiles{
		dataKey:     tdesktopDataNameKey(tdesktopAccountDataName(accountIndex)),
		sessionData: sessionData,
		mapData:     mapData,
	}, nil
}

func writeTDesktopSessionFiles(tdataPath string, files tdesktopSessionFiles) error {
	if err := os.MkdirAll(tdataPath, tdesktopDirMode); err != nil {
		return err
	}

	keyDataPath := filepath.Join(tdataPath, tdesktopKeyDataName) + "s"
	if err := tdesktopWriteFile(keyDataPath, files.keyData); err != nil {
		return err
	}

	for _, account := range files.accounts {
		if err := writeTDesktopAccountFiles(tdataPath, account); err != nil {
			return err
		}
	}
	return nil
}

func writeTDesktopAccountFiles(tdataPath string, account tdesktopAccountFiles) error {
	sessionPath := filepath.Join(tdataPath, account.dataKey)
	if err := os.MkdirAll(sessionPath, tdesktopDirMode); err != nil {
		return err
	}

	sessionDataPath := sessionPath + "s"
	if err := tdesktopWriteFile(sessionDataPath, account.sessionData); err != nil {
		return err
	}

	mapDataPath := filepath.Join(sessionPath, tdesktopMapFileName) + "s"
	return tdesktopWriteFile(mapDataPath, account.mapData)
}

func buildTDesktopKeyData(salt []byte, passKey []byte, tdesktopKey []byte, accountCount int) ([]byte, error) {
	encryptedKey, err := tdesktopEncryptLocal(tdesktopKey, passKey)
	if err != nil {
		return nil, err
	}

	info := buildTDesktopInfo(accountCount)

	encryptedInfo, err := tdesktopEncryptPadded(info, tdesktopKey)
	if err != nil {
		return nil, err
	}

	return tdesktopBuffers(salt, encryptedKey, encryptedInfo), nil
}

func buildTDesktopInfo(accountCount int) []byte {
	infoSize := tdesktopKeyInfoSize(accountCount)
	result := make([]byte, 0, infoSize)
	result = appendUint32LE(result, uint32(infoSize))
	result = appendUint32BE(result, uint32(accountCount))
	for i := 0; i < accountCount; i++ {
		result = appendUint32BE(result, uint32(i))
	}
	result = appendUint32BE(result, tdesktopActiveAccountIndex)
	return result
}

func tdesktopKeyInfoSize(accountCount int) int {
	return tdesktopBufferHeaderSize + tdesktopBufferHeaderSize + accountCount*4 + 4
}

func buildTDesktopSessionData(session TDataSession, authKey []byte, tdesktopKey []byte) ([]byte, error) {
	body := buildTDesktopSessionBody(buildTDesktopAuthBlock(session, authKey))
	encrypted, err := tdesktopEncryptLocal(body, tdesktopKey)
	if err != nil {
		return nil, err
	}

	return tdesktopBuffer(encrypted), nil
}

func buildTDesktopSessionBody(authBlock []byte) []byte {
	body := make([]byte, 0, tdesktopBufferHeaderSize+tdesktopBufferSize(authBlock))
	body = appendUint32BE(body, tdesktopSessionMagic)
	body = append(body, tdesktopBuffer(authBlock)...)
	return body
}

func buildTDesktopAuthBlock(session TDataSession, authKey []byte) []byte {
	authBlock := make([]byte, 0, 8+8+4+4+4+authKeySize+4)
	authBlock = appendUint32BE(authBlock, tdesktopWideIDMarker)
	authBlock = appendUint32BE(authBlock, tdesktopWideIDMarker)
	authBlock = appendUint64BE(authBlock, session.UserID)
	authBlock = appendUint32BE(authBlock, session.DC)
	authBlock = appendUint32BE(authBlock, tdesktopSingleAccountCount)
	authBlock = appendUint32BE(authBlock, session.DC)
	authBlock = append(authBlock, authKey...)
	authBlock = appendUint32BE(authBlock, tdesktopNoDestroyedKeyCount)
	return authBlock
}

func buildTDesktopEmptyMapData(tdesktopKey []byte) ([]byte, error) {
	encryptedMap, err := tdesktopEncryptLocal(nil, tdesktopKey)
	if err != nil {
		return nil, err
	}

	return tdesktopBuffers(nil, nil, encryptedMap), nil
}

func decodeAuthKey(authKeyHex string) ([]byte, error) {
	value := strings.TrimSpace(authKeyHex)
	if strings.HasPrefix(value, "0x") || strings.HasPrefix(value, "0X") {
		value = value[2:]
	}
	if len(value) != authKeySize*2 {
		return nil, fmt.Errorf("authKey must be %d hex chars", authKeySize*2)
	}

	authKey, err := hex.DecodeString(value)
	if err != nil {
		return nil, fmt.Errorf("decode authKey: %w", err)
	}
	if len(authKey) != authKeySize {
		return nil, fmt.Errorf("authKey must be %d bytes", authKeySize)
	}

	return authKey, nil
}

func randomBytes(length int) ([]byte, error) {
	data := make([]byte, length)
	if _, err := rand.Read(data); err != nil {
		return nil, err
	}
	return data, nil
}

func tdesktopDataNameKey(dataName string) string {
	sum := md5.Sum([]byte(dataName))
	result := make([]byte, 0, sessionFileNameSize)
	const hexDigits = "0123456789ABCDEF"
	for _, b := range sum[:sessionFileNameSize/2] {
		result = append(result, hexDigits[b&0x0f])
		result = append(result, hexDigits[b>>4])
	}
	return string(result)
}

func tdesktopAccountDataName(accountIndex int) string {
	if accountIndex == 0 {
		return tdesktopDataName
	}
	return fmt.Sprintf("%s#%d", tdesktopDataName, accountIndex+1)
}

func tdesktopEncrypt(data []byte, authKey []byte) ([]byte, error) {
	if len(data)%aes.BlockSize != 0 {
		return nil, io.ErrUnexpectedEOF
	}

	dataHash := sha1.Sum(data)
	messageKey := dataHash[:messageKeySize]
	aesKey, aesIV, err := calcOldAESKey(authKey, messageKey, false)
	if err != nil {
		return nil, err
	}

	encryptedData, err := aesIGEEncrypt(data, aesKey, aesIV)
	if err != nil {
		return nil, err
	}

	result := make([]byte, 0, messageKeySize+len(encryptedData))
	result = append(result, messageKey...)
	result = append(result, encryptedData...)
	return result, nil
}

func tdesktopEncryptLocal(data []byte, authKey []byte) ([]byte, error) {
	payloadSize := tdesktopBufferHeaderSize + len(data)
	payload := make([]byte, 0, payloadSize)
	payload = appendUint32LE(payload, uint32(payloadSize))
	payload = append(payload, data...)
	return tdesktopEncryptPadded(payload, authKey)
}

func tdesktopEncryptPadded(data []byte, authKey []byte) ([]byte, error) {
	padded, err := padToAESBlock(data)
	if err != nil {
		return nil, err
	}
	return tdesktopEncrypt(padded, authKey)
}

func tdesktopWriteFile(fileName string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(fileName), tdesktopDirMode); err != nil {
		return err
	}
	return os.WriteFile(fileName, tdesktopFileData(data), tdesktopFileMode)
}

func tdesktopFileData(data []byte) []byte {
	var versionBytes [4]byte
	binary.LittleEndian.PutUint32(versionBytes[:], tdesktopFileVersion)

	fileData := make([]byte, 0, len(tdesktopFileMagic)+len(versionBytes)+len(data)+md5.Size)
	fileData = append(fileData, tdesktopFileMagic...)
	fileData = append(fileData, versionBytes[:]...)
	fileData = append(fileData, data...)
	hashValue := tdesktopFileHash(data, versionBytes[:])
	fileData = append(fileData, hashValue[:]...)

	return fileData
}

func writeTDesktopSessionZip(w io.Writer, files tdesktopSessionFiles) error {
	zipWriter := zip.NewWriter(w)

	writeErr := writeTDesktopSessionZipEntries(zipWriter, files)
	closeErr := zipWriter.Close()
	if writeErr != nil {
		return writeErr
	}
	return closeErr
}

func writeTDesktopSessionZipEntries(zipWriter *zip.Writer, files tdesktopSessionFiles) error {
	if err := writeGeneratedZipDirectory(zipWriter, tdesktopZipRootName); err != nil {
		return err
	}

	keyDataPath := path.Join(tdesktopZipRootName, tdesktopKeyDataName+"s")
	if err := writeGeneratedZipFile(zipWriter, keyDataPath, tdesktopFileData(files.keyData)); err != nil {
		return err
	}

	for _, account := range files.accounts {
		if err := writeTDesktopAccountZipEntries(zipWriter, account); err != nil {
			return err
		}
	}
	return nil
}

func writeTDesktopAccountZipEntries(zipWriter *zip.Writer, account tdesktopAccountFiles) error {
	accountPath := path.Join(tdesktopZipRootName, account.dataKey)
	if err := writeGeneratedZipDirectory(zipWriter, accountPath); err != nil {
		return err
	}

	sessionDataPath := accountPath + "s"
	if err := writeGeneratedZipFile(zipWriter, sessionDataPath, tdesktopFileData(account.sessionData)); err != nil {
		return err
	}

	mapDataPath := path.Join(accountPath, tdesktopMapFileName+"s")
	return writeGeneratedZipFile(zipWriter, mapDataPath, tdesktopFileData(account.mapData))
}

func writeGeneratedZipDirectory(zipWriter *zip.Writer, name string) error {
	if !strings.HasSuffix(name, "/") {
		name += "/"
	}

	header := &zip.FileHeader{
		Name:   name,
		Method: zip.Store,
	}
	header.SetMode(tdesktopDirMode | fs.ModeDir)
	_, err := zipWriter.CreateHeader(header)
	return err
}

func writeGeneratedZipFile(zipWriter *zip.Writer, name string, data []byte) error {
	header := &zip.FileHeader{
		Name:   name,
		Method: zip.Deflate,
	}
	header.SetMode(tdesktopFileMode)

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = writer.Write(data)
	return err
}

func tdesktopBuffer(data []byte) []byte {
	result := make([]byte, 0, tdesktopBufferSize(data))
	return appendTDesktopBuffer(result, data)
}

func tdesktopBuffers(buffers ...[]byte) []byte {
	result := make([]byte, 0, tdesktopBuffersSize(buffers...))
	for _, buffer := range buffers {
		result = appendTDesktopBuffer(result, buffer)
	}
	return result
}

func tdesktopBufferSize(data []byte) int {
	return tdesktopBufferHeaderSize + len(data)
}

func tdesktopBuffersSize(buffers ...[]byte) int {
	size := 0
	for _, buffer := range buffers {
		size += tdesktopBufferSize(buffer)
	}
	return size
}

func appendTDesktopBuffer(dst []byte, data []byte) []byte {
	dst = appendUint32BE(dst, uint32(len(data)))
	return append(dst, data...)
}

func appendUint32BE(dst []byte, value uint32) []byte {
	var data [4]byte
	binary.BigEndian.PutUint32(data[:], value)
	return append(dst, data[:]...)
}

func appendUint64BE(dst []byte, value uint64) []byte {
	var data [8]byte
	binary.BigEndian.PutUint64(data[:], value)
	return append(dst, data[:]...)
}

func appendUint32LE(dst []byte, value uint32) []byte {
	var data [4]byte
	binary.LittleEndian.PutUint32(data[:], value)
	return append(dst, data[:]...)
}

func padToAESBlock(data []byte) ([]byte, error) {
	padding := aes.BlockSize - len(data)%aes.BlockSize
	if padding == aes.BlockSize {
		return data, nil
	}

	result := make([]byte, len(data)+padding)
	copy(result, data)
	if _, err := rand.Read(result[len(data):]); err != nil {
		return nil, err
	}
	return result, nil
}

func aesIGEEncrypt(plaintext []byte, key []byte, iv []byte) ([]byte, error) {
	if len(iv) != aes.BlockSize*2 {
		return nil, io.ErrUnexpectedEOF
	}
	if len(plaintext)%aes.BlockSize != 0 {
		return nil, io.ErrUnexpectedEOF
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	cipherText := make([]byte, len(plaintext))
	prevCipher := append([]byte(nil), iv[:aes.BlockSize]...)
	prevPlain := append([]byte(nil), iv[aes.BlockSize:]...)
	tmp := make([]byte, aes.BlockSize)
	cipherBlock := make([]byte, aes.BlockSize)

	for offset := 0; offset < len(plaintext); offset += aes.BlockSize {
		plainBlock := plaintext[offset : offset+aes.BlockSize]

		xorBlock(tmp, plainBlock, prevCipher)
		block.Encrypt(tmp, tmp)
		xorBlock(cipherBlock, tmp, prevPlain)

		copy(cipherText[offset:offset+aes.BlockSize], cipherBlock)
		copy(prevPlain, plainBlock)
		copy(prevCipher, cipherBlock)
	}

	return cipherText, nil
}

func tdesktopPassKey(salt []byte) []byte {
	passwordHash := sha512Bytes(salt, salt)
	return pbkdf2SHA512(passwordHash, salt)
}

func pbkdf2SHA512(password []byte, salt []byte) []byte {
	derivedKey := make([]byte, 0, authKeySize)
	var blockIndex [4]byte

	for block := uint32(1); len(derivedKey) < authKeySize; block++ {
		binary.BigEndian.PutUint32(blockIndex[:], block)

		mac := hmac.New(sha512.New, password)
		mac.Write(salt)
		mac.Write(blockIndex[:])
		derivedKey = mac.Sum(derivedKey)
	}

	return derivedKey[:authKeySize]
}

func validateTDesktopDC(dcID uint32) error {
	switch dcID {
	case 1, 2, 3, 4, 5:
		return nil
	default:
		return errors.New("Invalid DC")
	}
}
