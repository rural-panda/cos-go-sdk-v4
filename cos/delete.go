package cos

func (bucket *Bucket) DeleteObject(key string) error {
	_, err := bucket.client.do(deleteObject, bucket.BucketName, key, nil, "")
	if err != nil {
		return err
	}
	return nil
}