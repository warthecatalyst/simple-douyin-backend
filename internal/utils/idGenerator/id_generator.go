package idGenerator

import "github.com/google/uuid"

func GenerateVideoId() int64 {
	return int64(uuid.New().ID())
}

func GenerateUserId() int64 {
	return int64(uuid.New().ID())
}

func GenerateMessageId() int64 {
	return int64(uuid.New().ID())
}
