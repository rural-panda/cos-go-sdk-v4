package cos

import (
	"encoding/xml"
	"bytes"
	"net/url"
)

type deleteXML struct {
	XMLName xml.Name       `xml:"Delete"`
	Objects []DeleteObject `xml:"Object"` // 删除的所有Object
	Quiet   bool           `xml:"Quiet"`  // 安静响应模式
}

type DeleteObject struct {
	XMLName xml.Name `xml:"Object"`
	Key     string   `xml:"Key"`
}

// 批量删除的返回结果
type DeleteObjectsResult struct {
	XMLName        xml.Name `xml:"DeleteResult"`
	DeletedObjects []string `xml:"Deleted>Key"` // 删除的Object列表
}

func (bucket *Bucket) DeleteObject(key string) error {
	_, err := bucket.client.do(deleteObject, bucket.BucketName, key, nil, "")
	if err != nil {
		return err
	}
	return nil
}

func (bucket *Bucket) DeleteObjects(objectKeys []string)(DeleteObjectsResult, error) {
	dxml := deleteXML{}
	out := DeleteObjectsResult{}

	for _, key := range objectKeys {
		dxml.Objects = append(dxml.Objects, DeleteObject{Key: key})
	}
	dxml.Quiet = false
	bs, err := xml.Marshal(dxml)
	if err != nil {
		return out, err
	}
	buffer := new(bytes.Buffer)
	_, err = buffer.Write(bs)
	if err != nil {
		return out, err
	}

	res, err := bucket.client.do(multipleDeleteObject, bucket.BucketName, "", buffer, "")
	if err != nil {
		return out, err
	}
	defer res.Body.Close()

	if !dxml.Quiet {
		if err = xmlUnmarshal(res.Body, &out); err == nil {
			err = decodeDeleteObjectsResult(&out)
		}
	}
	return out, err
}

func decodeDeleteObjectsResult(result *DeleteObjectsResult) error {
	var err error
	for i := 0; i < len(result.DeletedObjects); i++ {
		result.DeletedObjects[i], err = url.QueryUnescape(result.DeletedObjects[i])
		if err != nil {
			return err
		}
	}
	return nil
}
