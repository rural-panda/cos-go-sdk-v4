package cos

import (
	"io"
	"crypto/sha1"
	"encoding/hex"
)

func calSha1FromReader(data io.Reader) string {
	sha := sha1.New()
	io.Copy(sha, data)
	return 	hex.EncodeToString(sha.Sum(nil))
}