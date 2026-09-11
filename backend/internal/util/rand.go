package util

import (
	"crypto/rand"
	"encoding/hex"
)

// RandHex 返回 n 字节的随机十六进制字符串。
func RandHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}
