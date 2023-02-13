package main

import (
	"context"
	"github.com/YOJIA-yukino/simple-douyin-backend/api"
	pb "github.com/YOJIA-yukino/simple-douyin-backend/api/rpc_controller_service/favorite/route"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"strconv"
	"time"
)

const (
	address = "localhost:50051"
)

func initRouter() {
	hServer := server.Default(server.WithHostPorts("127.0.0.1:8888"))
	hServer.POST("/douyin/favorite/action/", FavoriteAction)
	hServer.Spin()
}

// FavoriteAction 视频点赞接口
func FavoriteAction(ctx context.Context, requestContext *app.RequestContext) {
	token := requestContext.Query("token")
	videoIdstr := requestContext.Query("video_id")
	actionTypestr := requestContext.Query("action_type")
	userId, _ := strconv.ParseInt(token, 10, 64)
	videoId, _ := strconv.ParseInt(videoIdstr, 10, 64)
	actionType, _ := strconv.ParseInt(actionTypestr, 10, 32)
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	grpcClient := pb.NewFavoriteInfoClient(conn)
	contextIns, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := grpcClient.FavoriteAction(contextIns, &pb.FavoriteAction{
		UserId:     userId,
		VideoId:    videoId,
		ActionType: int32(actionType),
	})
	if err != nil {
		log.Fatalf("could not send: %v", err)
	}
	requestContext.JSON(consts.StatusOK, api.Response{StatusCode: r.StatusCode,
		StatusMsg: r.StatusMsg})
}

type VideoListResponse struct {
	api.Response
	VideoList []api.Video `json:"video_list"`
}

// FavoriteList 视频点赞列表
func FavoriteList(ctx context.Context, requestContext *app.RequestContext) {
	token := requestContext.Query("token")
	queryUserStr := requestContext.Query("user_id")
	loginUserId, _ := strconv.ParseInt(token, 10, 64)
	queryUserId, _ := strconv.ParseInt(queryUserStr, 10, 64)
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	grpcClient := pb.NewFavoriteInfoClient(conn)
	contextIns, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := grpcClient.FavoriteList(contextIns, &pb.UserFavorite{
		LoginUserId: loginUserId,
		QueryUserId: queryUserId,
	})
	if err != nil {
		log.Fatalf("could not send: %v", err)
	}
	success := false
	videoList := make([]api.Video, 0)
	for {
		v, err := stream.Recv()
		if err == io.EOF {
			success = true
			break
		}
		if err != nil {
			log.Fatalf("client.ListFeatures failed: %v", err)
		}
		videoList = append(videoList, api.Video{
			Id: v.Id,
			Author: api.User{
				Id:            v.Author.Id,
				Name:          v.Author.Name,
				FollowCount:   v.Author.FollowCount,
				FollowerCount: v.Author.FollowerCount,
				IsFollow:      v.Author.IsFollow,
			},
			PlayUrl:       v.PlayURL,
			CoverUrl:      v.CoverURL,
			FavoriteCount: v.FavoriteCount,
			CommentCount:  v.CommentCount,
			IsFavorite:    v.IsFavorite,
		})
	}
	if success {
		requestContext.JSON(consts.StatusOK, VideoListResponse{
			Response: api.Response{
				StatusCode: 0,
				StatusMsg:  "",
			},
			VideoList: videoList,
		})
	} else {
		requestContext.JSON(consts.StatusOK, VideoListResponse{
			Response: api.Response{
				StatusCode: int32(api.RecordNotExistErr),
				StatusMsg:  api.ErrorCodeToMsg[api.RecordNotExistErr],
			},
			VideoList: videoList,
		})
	}
}

func main() {
	initRouter()
}
