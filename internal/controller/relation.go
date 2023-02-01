package controller

import (
	"context"
	"github.com/YOJIA-yukino/simple-douyin-backend/api"
	"github.com/cloudwego/hertz/pkg/app"
	"net/http"
)

type UserListResponse struct {
	api.Response
	UserList []api.User `json:"user_list"`
}

// RelationAction no practical effect, just check if token is valid
func RelationAction(c context.Context, ctx *app.RequestContext) {
	token := ctx.Query("token")

	if _, exist := usersLoginInfo[token]; exist {
		ctx.JSON(http.StatusOK, api.Response{StatusCode: 0})
	} else {
		ctx.JSON(http.StatusOK, api.Response{StatusCode: 1, StatusMsg: "User doesn't exist"})
	}
}

// FollowList all users have same follow list
func FollowList(c context.Context, ctx *app.RequestContext) {
	ctx.JSON(http.StatusOK, UserListResponse{
		Response: api.Response{
			StatusCode: 0,
		},
		UserList: []api.User{DemoUser},
	})
}

// FollowerList all users have same follower list
func FollowerList(c context.Context, ctx *app.RequestContext) {
	ctx.JSON(http.StatusOK, UserListResponse{
		Response: api.Response{
			StatusCode: 0,
		},
		UserList: []api.User{DemoUser},
	})
}

// FriendList all users have same friend list
func FriendList(c context.Context, ctx *app.RequestContext) {
	ctx.JSON(http.StatusOK, UserListResponse{
		Response: api.Response{
			StatusCode: 0,
		},
		UserList: []api.User{DemoUser},
	})
}
