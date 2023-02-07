package router

import (
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/controller"
	"github.com/YOJIA-yukino/simple-douyin-backend/internal/utils/jwt"
	"github.com/cloudwego/hertz/pkg/app/server"
)

// InitRouter 初始化hertz服务器路由
func InitRouter(hertz *server.Hertz) {
	// public directory is used to serve static resources
	hertz.Static("/static", "./public")

	// 用户注册与登录需要进行鉴权, Feed可授权可不授权
	hertz.POST("/douyin/user/register", controller.Register)
	hertz.POST("/douyin/user/login", jwt.JwtMiddleware.LoginHandler)
	hertz.GET("/douyin/feed/", controller.Feed)

	// 鉴权authorization
	auth := hertz.Group("/douyin", jwt.JwtMiddleware.MiddlewareFunc())

	// basic apis
	auth.GET("/user/", controller.UserInfo)
	auth.POST("/publish/action/", controller.Publish)
	auth.GET("/publish/list/", controller.PublishList)

	// extra apis - I
	auth.POST("/favorite/action/", controller.FavoriteAction)
	auth.GET("/favorite/list/", controller.FavoriteList)
	auth.POST("/comment/action/", controller.CommentAction)
	auth.GET("/comment/list/", controller.CommentList)

	// extra apis - II
	auth.POST("/relation/action/", controller.RelationAction)
	auth.GET("/relation/follow/list/", controller.FollowList)
	auth.GET("/relation/follower/list/", controller.FollowerList)
	auth.GET("/relation/friend/list/", controller.FriendList)
	auth.GET("/message/chat/", controller.MessageChat)
	auth.POST("/message/action/", controller.MessageAction)
}
