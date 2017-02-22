package cos

import "fmt"

type ServiceError struct {
	Code    string   `json:"code"`
	Message string   `json:"message"`
	Data    string   `json:"data"`
}

func (e ServiceError) Error() string {
	return fmt.Sprintf("oss: service returned error: Code=%d, Message=%s, Data=%s",
		e.Code, e.Message, e.Data)
}