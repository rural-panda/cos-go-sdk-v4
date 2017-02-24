package cos

import (
	"fmt"
	"net/http"
	"encoding/xml"
	"encoding/json"
	"errors"
)

type ServiceErrorXML struct {
	Code      string   `xml:"Code"`
	Message   string   `xml:"Message"`
	TraceId   string   `xml:"TraceId"`
	Resource  string   `xml:"Resource"`
	RequestId string   `xml:"RequestId"`
}

// FIXME: 这个是遗弃的一套错误格式，腾讯云正在升级
type ServiceErrorJSON struct {
	Code    int64   `json:"errorcode"`
	Message string   `json:"errormsg"`
	RetCode int64   `json:"retcode"`
}

func (e ServiceErrorXML) Error() string {
	return fmt.Sprintf("Cos: service returned error: Code = %s, Message = %s, TraceID = %s, Resource = %s, RequestID = %s",
		e.Code, e.Message, e.TraceId, e.Resource, e.RequestId)
}

func (e ServiceErrorJSON) Error() string {
	return fmt.Sprintf("Cos: service returned error: Code = %d, Message = %s, RetCode = %d",
		e.Code, e.Message, e.RetCode)
}


func parseCosErr(resp *http.Response, body []byte) (error, error) {
	if resp.Header.Get("Content-Type") == "application/xml" || identifyXML(string(body)){
		var cosErr ServiceErrorXML
		if err := xml.Unmarshal(body, &cosErr); err != nil {
			return cosErr, errors.New("parse body to xml error: " + err.Error() +"  data: " + string(body))
		}
		return cosErr, nil
	}

	// 腾讯云可能返回 text/octet 的类型
	if resp.Header.Get("Content-Type") == "application/json" || identifyJSON(string(body)){
		var cosErr ServiceErrorJSON
		if err := json.Unmarshal(body, &cosErr); err != nil {
			return cosErr, errors.New("parse body to json error: " + err.Error() +"  data: " + string(body))
		}
		return cosErr, nil
	}
	return errors.New("not support error body format: " + resp.Header.Get("Content-Type") + " data: " + string(body)), nil
}