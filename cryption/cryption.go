package cryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"encoding/base64"
	"strings"
)

type SString struct {
	DummyChars []byte
	_internal  string
}

func (s *SString) DString() string {
	if strings.TrimSpace(s._internal) == "" || len(s.DummyChars) == 0 {
		return s._internal
	}
	stringFromBase64, err := base64.StdEncoding.DecodeString(s._internal)
	if err != nil {
		return s._internal
	}
	decryptBytes, err := DecryptBytes(stringFromBase64, s.DummyChars, []byte("0000000000000000"))
	if err != nil {
		return string(stringFromBase64)
	}
	return string(decryptBytes)
}

func (s *SString) SetDString(value string) {
	if strings.TrimSpace(value) == "" || len(s.DummyChars) == 0 {
		s._internal = value
	} else {
		encryptedBytes, _ := EncryptBytes([]byte(value), s.DummyChars, []byte("0000000000000000"))
		s._internal = base64.StdEncoding.EncodeToString(encryptedBytes)
	}
}

func (s *SString) CString() string {
	return s._internal
}

func (s *SString) SetCString(value string) {
	s._internal = value
}

func EncryptBytes(input, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	ciphertext := make([]byte, len(input))
	cfb.XORKeyStream(ciphertext, input)
	return ciphertext, nil
}

func DecryptBytes(ciphertext, key, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	cfb := cipher.NewCFBDecrypter(block, iv)
	plaintext := make([]byte, len(ciphertext))
	cfb.XORKeyStream(plaintext, ciphertext)
	return plaintext, nil
}

func PassphraseToDefaultKeyAndIV(data, salt []byte, count int) ([]byte, []byte) {
	hashList := make([]byte, 0)
	preHash := append(data, salt...)
	currentHash := md5.Sum(preHash)
	hashList = append(hashList, currentHash[:]...)

	for len(hashList) < 48 {
		preHash = append(currentHash[:], data...)
		preHash = append(preHash, salt...)
		currentHash = md5.Sum(preHash)
		hashList = append(hashList, currentHash[:]...)
	}

	key := hashList[:32]
	iv := hashList[32:48]
	return key, iv
}
