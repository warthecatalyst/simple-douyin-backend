package md5

import (
	"crypto/md5"
	"encoding/hex"
)

// MD5 use md5 to encrypt strings(like passwords)
func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
