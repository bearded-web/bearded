package services

import (
	"fmt"
	"net/http"

	"github.com/emicklei/go-restful"
)

type CodeErr int

const (
	CodeApp CodeErr = 1 // just application error
	// error codes related to db
	CodeDb        CodeErr = 18
	CodeIdHex     CodeErr = 19
	CodeDuplicate CodeErr = 20

	// Bad Request
	CodeWrongData   CodeErr = 40
	CodeWrongEntity CodeErr = 41

	// error codes related to auth
	CodeAuthReq    CodeErr = 60
	CodeAuthFailed CodeErr = 61
	CodeAuthForbid CodeErr = 62
)

var (
	AppErr         = NewError(CodeApp, "application error")
	DbErr          = NewError(CodeDb, "db error")
	IdHexErr       = NewError(CodeIdHex, "id should be bson uuid in hex form")
	WrongEntityErr = NewError(CodeWrongEntity, "wrong entity")
	DuplicateErr   = NewError(CodeDuplicate, "object with the same indexes is existed")
	AuthReqErr     = NewError(CodeAuthReq, "authorization required")
	AuthFailedErr  = NewError(CodeAuthFailed, "authorization failed")
	AuthForbidErr  = NewError(CodeAuthForbid, "you have no permission to this resource")
)

func NewError(c CodeErr, msg string) restful.ServiceError {
	return restful.NewError(int(c), msg)
}

func NewBadReq(msg string, args ...interface{}) restful.ServiceError {
	return NewError(CodeWrongData, fmt.Sprintf(msg, args...))
}

func NewAppErr(msg string) restful.ServiceError {
	return NewError(CodeApp, msg)
}

type ErrResp struct {
	Code int
	Err  error
}

func (e *ErrResp) Write(rw *restful.Response) {
	code := e.Code
	if code == 0 {
		code = http.StatusInternalServerError
	}
	if sErr, casted := e.Err.(restful.ServiceError); casted {
		rw.WriteServiceError(code, sErr)
	} else {
		rw.WriteError(code, e.Err)
	}
}

func (e *ErrResp) Error() string {
	code := e.Code
	if code == 0 {
		code = http.StatusInternalServerError
	}
	errMsg := ""
	if e.Err != nil {
		errMsg = e.Err.Error()
	}
	return fmt.Sprintf("%d %s: %s", http.StatusBadRequest, http.StatusText(http.StatusBadRequest), errMsg)
}
