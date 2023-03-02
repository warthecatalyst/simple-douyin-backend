package constants

import (
	"errors"
	"github.com/YOJIA-yukino/simple-douyin-backend/api"
)

var (
	InvalidTokenErr      = errors.New(api.ErrorCodeToMsg[api.TokenInvalidErr])
	NoVideoErr           = errors.New(api.ErrorCodeToMsg[api.NoVideoErr])
	UnKnownActionTypeErr = errors.New(api.ErrorCodeToMsg[api.UnKnownActionType])

	UserNotExistErr       = errors.New(api.ErrorCodeToMsg[api.UserNotExistErr])
	UserAlreadyExistErr   = errors.New(api.ErrorCodeToMsg[api.UserAlreadyExistErr])
	RecordNotExistErr     = errors.New(api.ErrorCodeToMsg[api.RecordNotExistErr])
	RecordAlreadyExistErr = errors.New(api.ErrorCodeToMsg[api.RecordAlreadyExistErr])
	RecordNotMatchErr     = errors.New(api.ErrorCodeToMsg[api.RecordNotMatchErr])
	InnerDataBaseErr      = errors.New(api.ErrorCodeToMsg[api.InnerDataBaseErr])
	InnerConnectionErr    = errors.New(api.ErrorCodeToMsg[api.InnerConnectionErr])
	RedisDBErr            = errors.New(api.ErrorCodeToMsg[api.RedisDBErr])
	KafkaServerErr        = errors.New(api.ErrorCodeToMsg[api.KafkaServerErr])
	KafkaClientErr        = errors.New(api.ErrorCodeToMsg[api.KafkaClientErr])
	CreateDataErr         = errors.New(api.ErrorCodeToMsg[api.CreateDataErr])

	VideoFormatErr = errors.New(api.ErrorCodeToMsg[api.VideoFormationErr])
	VideoSizeErr   = errors.New(api.ErrorCodeToMsg[api.VideoSizeErr])
	SavingFailErr  = errors.New(api.ErrorCodeToMsg[api.SavingFailErr])
	UploadFailErr  = errors.New(api.ErrorCodeToMsg[api.UploadFailErr])
	
	LockFailedErr = errors.New("lock Failed")
	TimeOutErr    = errors.New("timeout Error")
)
