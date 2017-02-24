package cos

import (
	"io"
	"strings"
	"net/http"
	"io/ioutil"
	"bytes"
	"fmt"
	"net/url"
	"strconv"
	"mime/multipart"
	"errors"
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


func (conn *cos) do(action, bucketName, objectName string, data io.Reader, standard string) (*Response, error) {
	var uri, method string
	objectName = url.QueryEscape(objectName)

	if action == headObject {
		uri = "/" + objectName
		method = "HEAD"
	} else if action == simpleUploadObject {
		uri = "/" + objectName
		method = "PUT"
	} else if action == simpleDownloadObject {
		uri = "/" + objectName
		method = "GET"
	} else if action == deleteObject {
		uri = "/" + objectName
		method = "DELETE"
	} else {
		return nil, fmt.Errorf("%s is not supported", action)
	}

	_url, err := parseURL(conn.config.SrvConf.Region + uri)
	if err != nil {
		return nil, fmt.Errorf("parse url error: %s", err)
	}
	return conn.doRequest(method, action, bucketName, objectName, _url, data, standard)
}

func parseURL(uri string) (*url.URL, error) {
	res, err := url.Parse(uri)
	return res, err
}

func (conn *cos) doRequest(method, action, bucket, object string, _url *url.URL, data io.Reader, standard string) (*Response, error) {
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
	req.Header.Set(HTTPHeaderHost, _url.Host)

	if action == simpleUploadObject {
		if standard == "" {
			req.Header.Set(StorageClass, DefaultStandard)
		} else {
			req.Header.Set(StorageClass, standard)
		}
		err := addContentLengthHeader(req, data)
		if err != nil {
			return nil, err
		}
		req.Header.Del("Transfer-Encoding")
	}


	conn.signXMLHeader(req, bucket, object, action)
	resp, err := conn.client.Do(req)
	if err != nil {
		return nil, errors.New("send request error : " + err.Error())
	}
	return conn.handleResponse(resp)
}

// json-api
func addDeleteObjectHeaderJSON(req *http.Request) {
	jsonStr := "{\"op\":\"delete\"}"
	req.Header.Set("Content-Type", "application/json")
	req.Body = ioutil.NopCloser(strings.NewReader(jsonStr))
	req.Header.Set("Content-Length", strconv.Itoa(len(jsonStr)))
}


// xml-api
func addContentLengthHeader(req *http.Request, data io.Reader) error {
	dataBytes, err := ioutil.ReadAll(data)
	if err != nil {
		return err
	}
	req.ContentLength = int64(len(dataBytes))
	data = bytes.NewBuffer(dataBytes)

	req.Body = ioutil.NopCloser(data)
	return nil
}


 //json api
func addUploadObjectHeaderJSON(req *http.Request, object string, data io.Reader) error {
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

	if statusCode >= 200 && statusCode <= 299 {
		return &Response{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			Body:       resp.Body,
		}, nil
	} else{
		bodyBytes, err := readResponseBody(resp)
		if err != nil {
			return nil, err
		}

		if len(bodyBytes) == 0 {
			err = fmt.Errorf("cos: service returned without a response body (%s)", resp.Status)
		} else {
			srvErr, errIn := parseCosErr(resp, bodyBytes)
			if errIn != nil {
				err = errIn
			} else {
				err = srvErr
			}
		}
		return &Response{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			Body:       ioutil.NopCloser(bytes.NewReader(bodyBytes)),
		}, err
	}
}


func readResponseBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	out, err := ioutil.ReadAll(resp.Body)
	if err == io.EOF {
		err = nil
	}
	return out, err
}
