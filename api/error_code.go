package api

//ErrorType 不同的error对应的errorCode，以及返回的message
type ErrorType int

const (
	UploadFailErr ErrorType = iota
	SavingFailErr
	VideoFormationErr
	VideoSizeErr
	NoVideoErr

	InnerDataBaseErr
	RedisDBErr
	KafkaServerErr
	KafkaClientErr
	CreateDataErr
	TokenInvalidErr
	UserNotExistErr
	UserAlreadyExistErr
	UserIdNotMatchErr
	RecordNotExistErr
	RecordAlreadyExistErr
	RecordNotMatchErr

	LogicErr
	UnKnownActionType
	InputFormatCheckErr
	GetDataErr
)

var ErrorCodeToMsg = map[ErrorType]string{
	UploadFailErr:     "Fail to upload File",
	SavingFailErr:     "Fail to save file",
	VideoFormationErr: "Video formation error",
	VideoSizeErr:      "Video size larger than expected",
	NoVideoErr:        "No video matches the requirement",

	InnerDataBaseErr:      "Inner database error",
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
