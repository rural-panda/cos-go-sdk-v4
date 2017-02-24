package cos

import (
	"strings"
	"errors"
)

type Bucket struct {
	client     *cos
	BucketName string
	standard   string
}

type ConfigHTTPTimeOut func(*HTTPTimeOut)

// oss 在这里进行了很多 url 的 parse 和 检测，主要是验证 endpoint 的格式，还有就是生成 client 时候，支持了代理
func New(appId, accessKeyID, accessKeySecret, region string, debug bool, act ConfigHTTPTimeOut) (*cos, error) {
	srvConf := &SrvConf{}
	srvConf.AppId = appId
	srvConf.SecretId = accessKeyID
	srvConf.SecretKey = accessKeySecret
	if !strings.HasPrefix(region, "http://") && !strings.HasPrefix(region, "https://") {
		region = "http://" + region
	}
	srvConf.Region = region

	httpTimeOut := getDefaultHttpTimeOutConfig()
	act(httpTimeOut)
	cosConfig := &Config{srvConf, httpTimeOut}

	cos := &cos{}
	cos.init(cosConfig, debug)
	return cos, nil
}

func (cos *cos) Bucket(bucket string) *Bucket {
	return &Bucket{cos, bucket, ""}
}

func (bucket *Bucket) SetStandard(standart string) error {
	if standart != Standard_IA && standart != Nearline && standart != Standard {
		return errors.New("not support strorage class : " + standart)
	}
	bucket.standard = standart
	return nil
}