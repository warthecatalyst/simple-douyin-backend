package controller

import (
	"context"
	"fmt"
	"github.com/YOJIA-yukino/simple-douyin-backend/api"
	"github.com/cloudwego/hertz/pkg/app"
	"net/http"
	"path/filepath"
)

type VideoListResponse struct {
	api.Response
	VideoList []api.Video `json:"video_list"`
}

// Publish check token then save upload file to public directory
func Publish(c context.Context, ctx *app.RequestContext) {
	token := ctx.PostForm("token")

	if _, exist := usersLoginInfo[token]; !exist {
		ctx.JSON(http.StatusOK, api.Response{StatusCode: 1, StatusMsg: "User doesn't exist"})
		return
	}

	data, err := ctx.FormFile("data")
	if err != nil {
		ctx.JSON(http.StatusOK, api.Response{
			StatusCode: 1,
			StatusMsg:  err.Error(),
		})
		return
	}

	filename := filepath.Base(data.Filename)
	user := usersLoginInfo[token]
	finalName := fmt.Sprintf("%d_%s", user.Id, filename)
	saveFile := filepath.Join("./public/", finalName)
	if err := ctx.SaveUploadedFile(data, saveFile); err != nil {
		ctx.JSON(http.StatusOK, api.Response{
			StatusCode: 1,
			StatusMsg:  err.Error(),
		})
		return
	}

	ctx.JSON(http.StatusOK, api.Response{
		StatusCode: 0,
		StatusMsg:  finalName + " uploaded successfully",
	})
}

// PublishList all users have same publish video list
func PublishList(c context.Context, ctx *app.RequestContext) {
	ctx.JSON(http.StatusOK, VideoListResponse{
		Response: api.Response{
			StatusCode: 0,
		},
		VideoList: DemoVideos,
	})
}
