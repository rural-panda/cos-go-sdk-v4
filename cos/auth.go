package cos

import (
	"strconv"
	"time"
	"strings"
	"crypto/hmac"
	"hash"
	"crypto/sha1"
	"io"
	"encoding/base64"
	"net/http"
	"math/rand"
	"net/url"
	"sort"
)

var (
	signValidity int64 = 60 * 15 // 每个请求的签名有效期为十五分钟
)


func (conn *cos) signXMLHeader(req *http.Request) string {
	srvConf := conn.config.SrvConf
	signedStr := "q-sign-algorithm=sha1&q-ak=" + srvConf.SecretId + "&q-sign-time="

	currentTime := strconv.FormatInt(time.Now().Unix(), 10)
	expiredTime := strconv.FormatInt(time.Now().Unix() + signValidity, 10)

	q_key_time := currentTime + ";" + expiredTime
	signedStr += q_key_time + "&q-key-time=" + q_key_time + "&q-header-list="

	sortedKeys, mapHeader := getSortedHeader(req)
	for i, v := range sortedKeys {
		if i > 0 {
			signedStr += ";"
		}
		signedStr += strings.ToLower(v)
	}
	signedStr += "&q-url-param-list=&q-signature="

	sign_key := hmacSha1(srvConf.SecretKey, q_key_time)
	// fixme: 这里需要做 query 部分的拼接
	formatStr := strings.ToLower(req.Method) + "\n" + req.URL.Path + "\n"  + "\n"
	//formatStr := strings.ToLower(req.Method) + "\n" + req.URL.Path + "\n" + url.QueryEscape(req.URL.RawQuery) + "\n"
	for i, v := range sortedKeys {
		if i > 0 {
			formatStr += "&"
		}
		formatStr += v + "=" + mapHeader[v]
	}
	formatStr += "\n"

	strToSign := "sha1\n" + q_key_time + "\n" + calStrSha1(formatStr) + "\n"

	signatureStr := hmacSha1(sign_key, strToSign)
	signedStr += signatureStr
	conn.Println("Authorization: " + signedStr)

	req.Header.Set(HTTPHeaderAuthorization, signedStr)
	return signedStr
}


func getSortedHeader(req *http.Request) ([]string, map[string]string) {
	headerMap := copyLowerAndURLEscapeHeader(req)
	tmpSlice := make([]string, 0, 20)

	for key, _ := range headerMap {
		tmpSlice = append(tmpSlice, key)
	}
	sort.Strings(tmpSlice)
	return tmpSlice, headerMap
}


// 因为 COS 签名需要小写 key，并且 value 要URL 编码，这里直接都做了
func copyLowerAndURLEscapeHeader(req *http.Request) map[string]string {
	headerMap := make(map[string]string)
	for key, value := range req.Header {
		headerMap[strings.ToLower(key)] = url.QueryEscape(sliceToStr(value))
	}
	if req.ContentLength != 0 {
		headerMap["content-length"] = strconv.FormatInt(req.ContentLength, 10)
	}
	return headerMap
}

// http.Header 对 Header 的值是以切片存储的，这里把他们又封装成一个字符串，以逗号分隔
func sliceToStr(src []string) string {
	str := ""
	for i, v := range src {
		if i > 0 {
			str += ","
		}
		str += v
	}
	return str
}

func (conn *cos) signJSONHeader(req *http.Request, bucket, key, action string) {
	srvConf := conn.config.SrvConf

	field := ""
	if action == deleteObject {
		fieldOriginal := "/" + srvConf.AppId + "/" + bucket + "/" + key
		field = base64.URLEncoding.EncodeToString([]byte(fieldOriginal))
	}

	currentTime := strconv.FormatInt(time.Now().Unix(), 10)
	expiredTime := strconv.FormatInt(time.Now().AddDate(0, 0, 1).Unix(), 10)
	randIntStr := strconv.FormatInt(int64(rand.Uint32()), 10)

	original := "a=appid&b=bucket&k=SecretID&e=expiredTime&t=currentTime&r=rand&f=fileid"
	signReplace := strings.NewReplacer("appid", srvConf.AppId, "bucket", bucket, "SecretID", srvConf.SecretId,
		"expiredTime", expiredTime, "currentTime", currentTime, "rand", randIntStr, "fileid", field)
	signStr := signReplace.Replace(original)

	h := hmac.New(func() hash.Hash {
		return sha1.New()
	}, []byte(srvConf.SecretKey))

	io.WriteString(h, signStr)
	signedTmp := h.Sum(nil)
	signedStr := base64.StdEncoding.EncodeToString([]byte(string(signedTmp) + signStr))

	req.Header.Set(HTTPHeaderAuthorization, signedStr)
}