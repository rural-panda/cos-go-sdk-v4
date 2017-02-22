package cos

import (
	"time"
)

type SrvConf struct {
	AppId     string
	SecretId  string
	SecretKey string
	Bucket    string
	Region    string
}

type HTTPTimeOut struct {
	ConnectTimeout   time.Duration
	ReadWriteTimeout time.Duration
	HeaderTimeout    time.Duration
	LongTimeout      time.Duration
}

type Config struct {
	SrvConf     *SrvConf
	HTTPTimeOut *HTTPTimeOut
}

func getDefaultHttpTimeOutConfig() *HTTPTimeOut {
	return &HTTPTimeOut{time.Second * 30, time.Second * 30, time.Second * 60, time.Second * 300}
}