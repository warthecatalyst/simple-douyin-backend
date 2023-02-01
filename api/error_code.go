package api

//ErrorType 不同的error对应的errorCode，以及返回的message
type ErrorType int

const (
	UploadFailErr ErrorType = iota
	SavingFailErr
	VideoFormationErr

	InnerDataBaseErr
	TokenInvalidErr
	UserNotExistErr
	UserAlreadyExistErr
	UserIdNotMatchErr
	UnKnownActionType
	RecordNotExistErr
	RecordExistErr

	LogicErr
	InputFormatCheckErr
)

var ErrorCodeToMsg = map[ErrorType]string{
	InnerDataBaseErr:    "发生数据库错误",
	SavingFailErr:       "存储文件错误",
	TokenInvalidErr:     "非法的Token",
	UserNotExistErr:     "用户不存在",
	UserAlreadyExistErr: "用户已存在",
	UserIdNotMatchErr:   "用户Id不匹配",
	UnKnownActionType:   "非法的操作",
	RecordNotExistErr:   "不存在对应的数据",
	UploadFailErr:       "文件上传失败",
	InputFormatCheckErr: "参数格式错误",
	LogicErr:            "逻辑错误",
	VideoFormationErr:   "上传视频有误",
	RecordExistErr:      "数据已存在",
}
