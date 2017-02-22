package cos

import (
	"io"
	"strings"
	"net/http"
	"time"
	"strconv"
	"crypto/hmac"
	"hash"
	"crypto/sha1"
	"encoding/base64"
	"io/ioutil"
	"bytes"
	"fmt"
	"mime/multipart"
	"net/textproto"
	"encoding/json"
	"math/rand"
	"net/url"
)

type Response struct {
	StatusCode int
	Headers    http.Header
	Body       io.ReadCloser
}

const (
	headObject = "headObject"
	deleteObject = "deleteObject"
	simpleUploadObject = "simpleUploadObject"
	simpleDownloadObject = "simpleDownloadObject"
)

const (
	urlPrefix = "/files/v2/"
	// COS 不同请求的 endpoint 可能不同
	urlPostfix = ".file.myqcloud.com"
)

func (conn *cos) do(action, bucketName, objectName string, data io.Reader) (*Response, error) {
	var uri, method string
	host := "http://"
	if action == headObject {
		uri = urlPrefix + conn.config.SrvConf.AppId + "/" + bucketName + "/" + objectName + "?op=stat"
		host += conn.config.SrvConf.Region + urlPostfix
		method = "GET"
	} else if action == simpleUploadObject {
		uri = urlPrefix + conn.config.SrvConf.Region + "/" + bucketName + "/" + objectName
		host += conn.config.SrvConf.Region + urlPostfix
		method = "POST"
	} else if action == simpleDownloadObject {
		uri = "/" + objectName
		host += bucketName + "-" + conn.config.SrvConf.AppId + "." + conn.config.SrvConf.Region + ".myqclound.com"
		method = "GET"
	} else if action == deleteObject {
		uri = urlPrefix + conn.config.SrvConf.AppId + "/" + bucketName + "/" + objectName
		host += conn.config.SrvConf.Region + urlPostfix
		method = "POST"
	} else {
		return nil, fmt.Errorf("%s is not supported", action)
	}
	uri = url.QueryEscape(uri)

	_url, err := parseURL(host + uri)
	if err != nil {
		return nil, fmt.Errorf("parse url error: %s", err)
	}
	return conn.doRequest(method, action, bucketName, objectName, _url, data)
}


func parseURL(uri string) (*url.URL, error) {
	res, err := url.Parse(uri)
	return res, err
}


func (conn *cos) doRequest(method, action, bucket, object string, _url *url.URL, data io.Reader) (*Response, error) {
	method = strings.ToUpper(method)
	req := &http.Request{
		Method:     method,
		URL:        _url,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     make(http.Header),
		Host:       _url.Host,
	}

	conn.signHeader(req, bucket, object, action)
	req.Header.Set(HTTPHeaderHost, _url.Host)

	if action == simpleDownloadObject {
		addDeleteObjectHeader(req)
	}
	if action == simpleUploadObject {
		err := addUploadObjectHeader(req, object, data)
		if err != nil {
			return nil, err
		}
	}

	resp, err := conn.client.Do(req)
	if err != nil {
		return nil, err
	}
	return conn.handleResponse(resp)
}


func addDeleteObjectHeader(req *http.Request) {
	jsonStr := "{\"op\":\"delete\"}"
	req.Header.Set("Content-Type", "application/json")
	req.Body = ioutil.NopCloser(strings.NewReader(jsonStr))
	req.Header.Set("Content-Length", strconv.Itoa(len(jsonStr)))
}


func addUploadObjectHeader(req *http.Request, object string, data io.Reader) error {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	file, err := writer.CreateFormFile("filecontent", object)
	if err != nil {
		return err
	}
	io.Copy(file, data)

	writer.WriteField("op", "upload")
	writer.WriteField("sha", calSha1FromReader(data))
	req.Header.Set("Content-Type", writer.FormDataContentType())

	if err := writer.Close(); err != nil {
		return err
	}
	req.Body = ioutil.NopCloser(body)
	return nil
}


func (conn *cos) handleResponse(resp *http.Response) (*Response, error) {
	statusCode := resp.StatusCode

	// 正确处理的请求，直接返回，让调用者处理
	if statusCode >= 200 && statusCode <= 299 {
		return &Response{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			Body:       resp.Body,
		}, nil
	}
	if statusCode >= 400 && statusCode <= 599 {
		var respBody []byte
		respBody, err := readResponseBody(resp)
		if err != nil {
			return nil, err
		}

		if len(respBody) == 0 {
			err = fmt.Errorf("cos: service returned without a response body (%s)", resp.Status)
		} else {
			// 正常的错误，按照 COS 的结构做解析
			srvErr, errIn := parseCosErr(respBody)
			if errIn != nil {
				err = errIn
			} else {
				err = srvErr
			}
		}
		return &Response{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			Body:       ioutil.NopCloser(bytes.NewReader(respBody)),
		}, err
	} else {
		// 3xx StatusCode and other cases
		err := fmt.Errorf("oss: service returned %d,%s", resp.StatusCode, resp.Status)
		return &Response{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			Body:       resp.Body,
		}, err
	}
}

func parseCosErr(body []byte) (ServiceError, error) {
	var cosErr ServiceError
	if err := json.Unmarshal(body, &cosErr); err != nil {
		return cosErr, err
	}
	return cosErr, nil
}

func readResponseBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	out, err := ioutil.ReadAll(resp.Body)
	if err == io.EOF {
		err = nil
	}
	return out, err
}

func (conn *cos) signHeader(req *http.Request, bucket, key, action string) {
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