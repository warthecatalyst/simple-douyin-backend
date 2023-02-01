package controller

import (
	"context"
	"github.com/YOJIA-yukino/simple-douyin-backend/api"
	"github.com/cloudwego/hertz/pkg/app"
	"net/http"
)

type CommentListResponse struct {
	api.Response
	CommentList []api.Comment `json:"comment_list,omitempty"`
}

type CommentActionResponse struct {
	api.Response
	Comment api.Comment `json:"comment,omitempty"`
}

// CommentAction no practical effect, just check if token is valid
func CommentAction(c context.Context, ctx *app.RequestContext) {
	token := ctx.Query("token")
	actionType := ctx.Query("action_type")

	if user, exist := usersLoginInfo[token]; exist {
		if actionType == "1" {
			text := ctx.Query("comment_text")
			ctx.JSON(http.StatusOK, CommentActionResponse{Response: api.Response{StatusCode: 0},
				Comment: api.Comment{
					Id:         1,
					User:       user,
					Content:    text,
					CreateDate: "05-01",
				}})
			return
		}
		ctx.JSON(http.StatusOK, api.Response{StatusCode: 0})
	} else {
		ctx.JSON(http.StatusOK, api.Response{StatusCode: 1, StatusMsg: "User doesn't exist"})
	}
}

// CommentList all videos have same demo comment list
func CommentList(c context.Context, ctx *app.RequestContext) {
	ctx.JSON(http.StatusOK, CommentListResponse{
		Response:    api.Response{StatusCode: 0},
		CommentList: DemoComments,
	})
}
