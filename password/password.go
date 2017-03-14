// GAME SERVER APPLICATION
//
// Copyright 2016 ,Authors [蒋程]. All rights reserved.
//
// This Source Code Form is subject to the terms of the 锦绣山河 License v1.0.
// 禁止非法复制。

// 这个包提供服务端加解密的基本方法。主要实现以下两种应用的支持：
// 1、用户密码加密、验证支持。使用base64和SHA256
// 2、字符串的对称加解密支持。使用cipher和des
package password

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"crypto/sha256"
	"encoding/base64"
	"math/rand"
	"time"
)

const (
	constRandomLength = 16
	constAsciiPad     = 31
)

var rnd = rand.New(rand.NewSource(time.Now().Unix()))

// 本方法生成一个由标准字符构成的随机字符串作为salt,其长度以输入参数决定。
// 需要注意的是，如果输入参数相同，但调用函数中有影响到rand，则结果不会相同。
//	@length:salt的string长度
//	@asciiPad:一个值介于[0,127）的设置值，用于决定salt每个字节的值的范围。asciiPad越大，salt各个字节的随机取值范围越小。建议asciiPad小于40。
func GenerateSalt(length, asciiPad int) string {
	var salt []byte
	if length < 0 {
		length = constRandomLength
	}
	if asciiPad < 0 && asciiPad >= 127 {
		asciiPad = constAsciiPad
	}
	//循环指定次数，每次增加一个字节，字节内容为[_ASCIIPAD,127]之间。
	for i := 0; i < length; i++ {
		salt = append(salt, byte(rnd.Intn(127-asciiPad)+asciiPad))
	}
	return string(salt)
}

// 根据输入的密码和salt，返还加密字符串。
// 主要用于密码验证。
func GenerateHash(salt string, password string) string {
	var hash string
	fullString := salt + password
	sha := sha256.New()
	sha.Write([]byte(fullString))

	hashCode := sha.Sum(nil)
	hash = base64.StdEncoding.EncodeToString(hashCode)

	return hash
}

// 将明文转换为随机的salt和hash。
// 主要用于密码生成。
func ReturnPassword(password string) (salt, hash string) {
	salt = GenerateSalt(8, 20)
	hash = GenerateHash(salt, password)
	return salt, hash
}

// DesEncrypt 利用对称加密算法，生成密码
//	@src:	明文待加密数据
//	@key:	为加解密的密钥，通常作为服务端的salt存在数据库
func DesEncrypt(src, key []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	src = pkcs5Padding(src, block.BlockSize())
	// src = ZeroPadding(src, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, key)
	crypted := make([]byte, len(src))
	// 根据CryptBlocks方法的说明，如下方式初始化crypted也可以
	// crypted := src
	blockMode.CryptBlocks(crypted, src)
	return crypted, nil

}

// DesDecrypt 利用对称解密算法，解密，将密文转换成明文
//	@dst:	加密数据
//	@key:	为加解密的密钥，通常作为服务端的salt存在数据库
func DesDecrypt(dst, key []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}

	blockMode := cipher.NewCBCDecrypter(block, key)
	origData := make([]byte, len(dst))

	// origData := dst

	blockMode.CryptBlocks(origData, dst)

	origData = pkcsS5UnPadding(origData)

	// origData = ZeroUnPadding(origData)

	return origData, nil

}

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)

}

func pkcsS5UnPadding(origData []byte) []byte {
	length := len(origData)
	// 去掉最后一个字节 unpadding 次

	unpadding := int(origData[length-1])

	return origData[:(length - unpadding)]

}
