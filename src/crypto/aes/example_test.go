package aes_test

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
)

func ExampleNewAES_encrypt() {
	key := []byte("m0M4Ga7DUQr8OwaXLUruuUEF6II8MK1U") // 32 bytes
	text := []byte("exampleplaintext")
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}
	b := base64.StdEncoding.EncodeToString(text)
	ciphertext := make([]byte, aes.BlockSize+len(b))
	// iv is 00 static for test purpose, generally it should be random
	iv := ciphertext[:aes.BlockSize]
	// uncomment following if you want random iv
	// if _, err := io.ReadFull(rand.Reader, iv); err != nil {
	// 	return nil, err
	// }
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext[aes.BlockSize:], []byte(b))
	fmt.Printf("%x\n", ciphertext)
	//Output: 0000000000000000000000000000000005168330cdce49197c690bd57f12385a2150f905a6ed463f
}

func ExampleNewAES_decrypt() {
	key := []byte("m0M4Ga7DUQr8OwaXLUruuUEF6II8MK1U") // 32 bytes
	text, err := hex.DecodeString("0000000000000000000000000000000005168330cdce49197c690bd57f12385a2150f905a6ed463f")
	if err != nil {
		panic(err.Error())
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}
	if len(text) < aes.BlockSize {
		panic(errors.New("ciphertext too short"))
	}
	iv := text[:aes.BlockSize]
	text = text[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(text, text)
	data, err := base64.StdEncoding.DecodeString(string(text))
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("%s\n", data)
	// Output: exampleplaintext
}
