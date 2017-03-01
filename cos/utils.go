package cos

import (
	"io"
	"crypto/sha1"
	"encoding/hex"
	"crypto/hmac"
	"hash"
	"fmt"
	"strings"
	"crypto/md5"
	"encoding/xml"
	"io/ioutil"
)

func calSha1FromReader(data io.Reader) string {
	sha := sha1.New()
	io.Copy(sha, data)
	return 	hex.EncodeToString(sha.Sum(nil))
}

func xmlUnmarshal(body io.Reader, v interface{}) error {
	data, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	return xml.Unmarshal(data, v)
}

func calStrSha1(str string) string {
	sha := sha1.New()
	sha.Write([]byte(str))
	return hex.EncodeToString(sha.Sum(nil))
}

func calStrMD5(str string) string {
	_md5 := md5.New()
	_md5.Write([]byte(str))
	return hex.EncodeToString(_md5.Sum(nil))
}


func (conn *cos) Println(a ...interface{}) {
	if conn.debug {
		fmt.Println(a)
		fmt.Println()
	}
}

// 因为有时候不返回 Content-Type 头
func identifyJSON(str string) bool {
	return strings.HasPrefix(str, "{") && strings.HasSuffix(str, "}")
}


func identifyXML(str string) bool {
	return strings.HasPrefix(str, "<") && strings.HasSuffix(str, ">")
}


func hmacSha1(key, str string) string {
	h := hmac.New(func() hash.Hash {
		return sha1.New()
	}, []byte(key))
	io.WriteString(h, str)
	return string(hex.EncodeToString(h.Sum(nil)))
}