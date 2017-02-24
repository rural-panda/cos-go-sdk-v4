package cos

import (
	"time"
)

const (
	Standard_IA = "Standard_IA"
	Nearline = "Nearline"
	Standard = "Standard"
	DefaultStandard = Standard_IA
)




type SrvConf struct {
	AppId     string
	SecretId  string
	SecretKey string
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
	return &HTTPTimeOut{time.Second * 30, time.Second * 30, time.Second * 30, time.Second *35}
}