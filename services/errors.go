package services

import "github.com/emicklei/go-restful"

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

func NewBadReq(msg string) restful.ServiceError {
	return NewError(CodeWrongData, msg)
}

func NewAppErr(msg string) restful.ServiceError {
	return NewError(CodeApp, msg)
}
