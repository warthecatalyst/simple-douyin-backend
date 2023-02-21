package service

import (
	"bytes"
	"context"
	"github.com/YOJIA-yukino/simple-douyin-backend/api"
	pbservice "github.com/YOJIA-yukino/simple-douyin-backend/api/rpc_controller_service/video"
	pbdao "github.com/YOJIA-yukino/simple-douyin-backend/api/rpc_service_dao/video"
	initialization "github.com/YOJIA-yukino/simple-douyin-backend/init"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/model"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/oss"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/constants"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/files"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/idGenerator"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"io"
	"path"
	"strconv"
	"sync"
	"time"
)

// videoService 与publish相关的操作集合
type videoService struct {
	pbservice.UnimplementedVideoServiceInfoServer
}

func getUploadPath(userId int64, fileName string) string {
	return initialization.OssConf.BucketDirectory + "/" + strconv.FormatInt(userId, 10) + "/" + fileName
}

// getUploadURL 得到一名用户对应的云端存储路径
func getUploadURL(userId int64, fileName string) string {
	return "https://" + initialization.OssConf.Bucket + "." + initialization.OssConf.Url + "/" + getUploadPath(userId, fileName)
}

var (
	publishServiceInstance *videoService
	publishOnce            sync.Once
)

const (
	userPublishPrefix     = "user_publish_"
	userPublishExpireTime = 90 * time.Minute
)

// GetVideoServiceInstance 获取publishServiceInstance的实例
func GetVideoServiceInstance() *videoService {
	initRedis()
	initKafka()
	publishOnce.Do(func() {
		publishServiceInstance = &videoService{}
	})
	return publishServiceInstance
}

func (p *videoService) PublishVideoInfo(ctx context.Context, in *pbservice.VideoServicePost) (*wrapperspb.BoolValue, error) {
	userId := in.UserId
	title := in.Title
	fileName := in.FileName
	fileSize := in.FileSize
	content := in.Content
	err := p.PublishInfo(&content, userId, fileSize, title, fileName)
	if err != nil {
		return &wrapperspb.BoolValue{Value: false}, err
	} else {
		return &wrapperspb.BoolValue{Value: true}, status.New(codes.OK, "").Err()
	}
}

func (p *videoService) uploadVideoToOSS(data *[]byte, userId int64, filename string) error {
	reader := bytes.NewReader(*data)

	// 先将文件流上传至BucketDirectory目录下
	err := oss.UploadFromReader(getUploadPath(userId, filename), reader)
	if err != nil {
		logger.GlobalLogger.Printf("Error in UploadFromReader: %v", err.Error())
		return err
	}

	return nil
}

func (p *videoService) uploadCoverToOSS(userId int64, filepath, filename string) error {
	if err := oss.UploadFromFile(getUploadPath(userId, filename), filepath); err != nil {
		logger.GlobalLogger.Printf("Error in UploadFromFile: %v", err.Error())
		return err
	}

	return nil
}

// PublishInfo service层上传user的一个视频
func (p *videoService) PublishInfo(data *[]byte, userId, fileSize int64, title, fileName string) error {
	logger.GlobalLogger.Printf("title = %v", title)
	logger.GlobalLogger.Printf("fileName = %v", fileName)
	//首先检查video的扩展名与大小
	if !files.CheckFileExt(fileName) {
		return constants.VideoFormatErr
	}
	if !files.CheckFileSize(fileSize) {
		return constants.VideoSizeErr
	}

	logger.GlobalLogger.Print("Start Saving")
	//然后将文件保存至本地
	saveDir := path.Join(initialization.VideoConf.SavePath, strconv.FormatInt(userId, 10))
	videoName, err := files.SaveDataToLocal(saveDir, data, fileName)
	if err != nil {
		logger.GlobalLogger.Printf("Time = %v, Saving Video Error = %v", time.Now(), err.Error())
		return constants.SavingFailErr
	}

	//截取视频的第一帧作为cover
	saveVideo := saveDir + "/" + videoName
	coverName := files.GetFileNameWithoutExt(videoName) + "_cover" + ".jpeg"
	saveCover := saveDir + "/" + coverName
	err = files.ExtractCoverFromVideo(saveVideo, saveCover)
	if err != nil {
		logger.GlobalLogger.Printf("Time = %v, Extracting Cover Error = %v", time.Now(), err.Error())
		return constants.SavingFailErr
	}

	//上传视频与封面
	logger.GlobalLogger.Print("Saving Complete, Start Uploading")
	err = p.uploadVideoToOSS(data, userId, videoName)
	if err != nil {
		logger.GlobalLogger.Printf("Time = %v, Extracting Cover Error = %v", time.Now(), err.Error())
		return constants.UploadFailErr
	}
	err = p.uploadCoverToOSS(userId, saveCover, coverName)

	//RPC写入数据库
	address := initialization.RpcSDConf.VideoServiceHost + initialization.RpcSDConf.VideoServicePort
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.GlobalLogger.Printf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pbdao.NewVideoDaoInfoClient(conn)
	videoId := idGenerator.GenerateVideoId()
	// 更新用户的publish list
	// Contact the server and print out its response.
	ctx1, cancel1 := context.WithTimeout(context.Background(), time.Second)
	defer cancel1()
	_, err = c.AddVideo(ctx1, &pbdao.VideoDaoPost{
		VideoId:   videoId,
		UserId:    userId,
		VideoName: title,
		PlayURL:   getUploadURL(userId, videoName),
		CoverURL:  getUploadURL(userId, coverName),
	})
	return err
}

// PublishListInfo service层获得用户userId所有发表过的视频
func (p *videoService) PublishListInfo(userId, loginUserId int64) ([]api.Video, error) {
	var err error
	//RPC写入数据库
	address := initialization.RpcSDConf.VideoServiceHost + initialization.RpcSDConf.VideoServicePort
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.GlobalLogger.Printf("did not connect: %v", err)
	}
	defer conn.Close()
	grpcClient := pbdao.NewVideoDaoInfoClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stream, err := grpcClient.GetPublishIdList(ctx, &wrapperspb.Int64Value{Value: userId})
	if err != nil {
		return nil, err
	}
	videoIdList := make([]int64, 0)
	for {
		videoResp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			logger.GlobalLogger.Printf("get Videos From Dao Failed")
			return nil, err
		}
		videoIdList = append(videoIdList, videoResp.Value)
	}
	go p.putPublishListInRedis(userId, videoIdList)
	videoList := make([]*model.Video, 0)

	apiVideos, err := getVideoListByModel(loginUserId, videoList)
	return apiVideos, nil
}

func (p *videoService) getVideoListThroughVideoIdList(videoIdList []int64) {
}

func (p *videoService) updatePublishListInRedis(userId, videoId int64) {
}

func (p *videoService) putPublishListInRedis(userId int64, videoIds []int64) {

}
