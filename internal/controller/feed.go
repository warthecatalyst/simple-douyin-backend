package controller

import (
	"context"
	"github.com/YOJIA-yukino/simple-douyin-backend/api"
	"github.com/cloudwego/hertz/pkg/app"
	"net/http"
	"time"
)

type FeedResponse struct {
	api.Response
	VideoList []api.Video `json:"video_list,omitempty"`
	NextTime  int64       `json:"next_time,omitempty"`
}

// Feed same demo video list for every request
func Feed(c context.Context, ctx *app.RequestContext) {
	ctx.JSON(http.StatusOK, FeedResponse{
		Response:  api.Response{StatusCode: 0},
		VideoList: DemoVideos,
		NextTime:  time.Now().Unix(),
	})
}
