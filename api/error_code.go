package api

//ErrorType 不同的error对应的errorCode，以及返回的message
type ErrorType int

const (
	UploadFailErr     ErrorType = 10001
	SavingFailErr     ErrorType = 10002
	VideoFormationErr ErrorType = 10003
	VideoSizeErr      ErrorType = 10004
	NoVideoErr        ErrorType = 10005

	InnerDataBaseErr      ErrorType = 10101
	InnerConnectionErr    ErrorType = 10102
	RedisDBErr            ErrorType = 10103
	KafkaServerErr        ErrorType = 10104
	KafkaClientErr        ErrorType = 10105
	CreateDataErr         ErrorType = 10106
	TokenInvalidErr       ErrorType = 10107
	UserNotExistErr       ErrorType = 10108
	UserAlreadyExistErr   ErrorType = 10109
	UserIdNotMatchErr     ErrorType = 10110
	RecordNotExistErr     ErrorType = 10111
	RecordAlreadyExistErr ErrorType = 10112
	RecordNotMatchErr     ErrorType = 10113

	LogicErr            ErrorType = 10201
	UnKnownActionType   ErrorType = 10202
	InputFormatCheckErr ErrorType = 10203
	GetDataErr          ErrorType = 10204
)

var ErrorCodeToMsg = map[ErrorType]string{
	UploadFailErr:     "Fail to upload File",
	SavingFailErr:     "Fail to save file",
	VideoFormationErr: "Video formation error",
	VideoSizeErr:      "Video size larger than expected",
	NoVideoErr:        "No video matches the requirement",

	InnerDataBaseErr:      "Inner database error",
	InnerConnectionErr:    "Inner Connection error",
	RedisDBErr:            "Redis Cache error",
	KafkaServerErr:        "Kafka Server error",
	KafkaClientErr:        "Kafka Client error",
	CreateDataErr:         "Create data error",
	TokenInvalidErr:       "Invalid Token",
	UserNotExistErr:       "用户名或密码错误",
	UserAlreadyExistErr:   "用户名已存在",
	UserIdNotMatchErr:     "Not match userId",
	RecordNotExistErr:     "Record does not exist",
	RecordAlreadyExistErr: "Record already exists",
	RecordNotMatchErr:     "Record doesn't match",

	LogicErr:            "Inner logic error",
	UnKnownActionType:   "Unknown Action Type",
	InputFormatCheckErr: "Input formation error",
	GetDataErr:          "Fail to get data from context",
}
