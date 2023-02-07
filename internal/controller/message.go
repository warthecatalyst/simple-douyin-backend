package controller

import (
	"context"
	"fmt"
	"github.com/YOJIA-yukino/simple-douyin-backend/api"
	"github.com/cloudwego/hertz/pkg/app"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

var tempChat = map[string][]api.Message{}

var messageIdSequence = int64(1)

type ChatResponse struct {
	api.Response
	MessageList []api.Message `json:"message_list"`
}

// MessageAction no practical effect, just check if token is valid
func MessageAction(c context.Context, ctx *app.RequestContext) {
	token := ctx.Query("token")
	toUserId := ctx.Query("to_user_id")
	content := ctx.Query("content")

	if user, exist := usersLoginInfo[token]; exist {
		userIdB, _ := strconv.Atoi(toUserId)
		chatKey := genChatKey(user.Id, int64(userIdB))

		atomic.AddInt64(&messageIdSequence, 1)
		curMessage := api.Message{
			Id:         messageIdSequence,
			Content:    content,
			CreateTime: time.Now().Format(time.Kitchen),
		}

		if messages, exist := tempChat[chatKey]; exist {
			tempChat[chatKey] = append(messages, curMessage)
		} else {
			tempChat[chatKey] = []api.Message{curMessage}
		}
		ctx.JSON(http.StatusOK, api.Response{StatusCode: 0})
	} else {
		ctx.JSON(http.StatusOK, api.Response{StatusCode: 1, StatusMsg: "User doesn't exist"})
	}
}

// MessageChat all users have same follow list
func MessageChat(c context.Context, ctx *app.RequestContext) {
	token := ctx.Query("token")
	toUserId := ctx.Query("to_user_id")

	if user, exist := usersLoginInfo[token]; exist {
		userIdB, _ := strconv.Atoi(toUserId)
		chatKey := genChatKey(user.Id, int64(userIdB))

		ctx.JSON(http.StatusOK, ChatResponse{Response: api.Response{StatusCode: 0}, MessageList: tempChat[chatKey]})
	} else {
		ctx.JSON(http.StatusOK, api.Response{StatusCode: 1, StatusMsg: "User doesn't exist"})
	}
}

func genChatKey(userIdA int64, userIdB int64) string {
	if userIdA > userIdB {
		return fmt.Sprintf("%d_%d", userIdB, userIdA)
	}
	return fmt.Sprintf("%d_%d", userIdA, userIdB)
}
