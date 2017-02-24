package cos

import (
	"net/http"
)

func (bucket *Bucket) HeadObject(key string) (http.Header, error) {
	response, err := bucket.client.do(headObject, bucket.BucketName, key, nil, "")
	if err != nil {
		return nil, err
	}

	return response.Headers, nil
}