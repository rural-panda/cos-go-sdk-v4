package cos

import "io"

func (bucket *Bucket) PutObject(key string, object io.Reader) error {
	_, err := bucket.client.do(simpleUploadObject, bucket.BucketName, key, object)
	if err != nil {
		return err
	}
	return nil
}