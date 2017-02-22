package cos

// oss 在这里进行了很多 url 的 parse 和 检测，主要是验证 endpoint 的格式，还有就是生成 client 时候，支持了代理
func New(appId, accessKeyID, accessKeySecret, region string) (*cos, error) {
	srvConf := &SrvConf{}
	srvConf.AppId = appId
	srvConf.SecretId = accessKeyID
	srvConf.SecretKey = accessKeySecret
	srvConf.Region = region

	httpTimeOut := getDefaultHttpTimeOutConfig()
	cosConfig := Config{srvConf, httpTimeOut}

	cos := &cos{}
	cos.init(cosConfig)
	return cos, nil
}


func (cos *cos) Bucket(bucket string) Bucket {
	return &Bucket{cos, bucket}
}