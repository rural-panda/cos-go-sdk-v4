package cos

import "io"

func (bucket *Bucket) GetObject(key string) (io.ReadCloser, error) {
	response, err := bucket.client.do(simpleDownloadObject, bucket.BucketName, key, nil, "")
	if err != nil {
		return nil, err
	}
	return response.Body, nil
}
