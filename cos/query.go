package cos

import (
	"net/http"
)

func (bucket *Bucket) HeadObject(key string) (http.Header, error) {
	response, err := bucket.client.do(headObject, bucket.BucketName, key, nil)
	if err != nil {
		return nil, err
	}
	header, err := responseToHttpHeader(response)
	if err != nil {
		return nil, err
	}
	return header, nil
}


func responseToHttpHeader(response Response) (http.Header, error) {
	return http.Header{}, nil
}


type headObjectResult struct {
	Access_url     string
	Authority      string
	Biz_attr       string
	CTime          string
	Custom_headers http.Header
	FileLen        int64
	FileSize       int64
	Forbid         int64
	MTime          int64
	Preview_url    string
	Sha            string
	SliceSize      int64
	Source_url     string
}