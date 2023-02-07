package model

import (
	"gorm.io/gorm"
	"time"
)

// Video 视频：数据库实体
type Video struct {
	gorm.Model
	VideoID       int64  `gorm:"type:BIGINT;not null;UNIQUE"`
	VideoName     string `gorm:"type:varchar(100);not null"`
	UserID        int64  `gorm:"type:BIGINT;not null;index:idx_author_id"`
	FavoriteCount int32  `gorm:"type:INT;not null;default:0"`
	CommentCount  int32  `gorm:"type:INT;not null;default:0"`
	PlayURL       string `gorm:"type:varchar(200);not null"`
	CoverURL      string `gorm:"type:varchar(200);not null"`
}

// User 用户:数据库实体
type User struct {
	gorm.Model
	UserID        int64  `gorm:"type:bigint;unsigned;not null;unique;uniqueIndex:idx_user_id" json:"user_id"`
	UserName      string `gorm:"type:varchar(50);not null;unique;uniqueIndex:idx_user_name" json:"name" validate:"min=6,max=32"`
	PassWord      string `gorm:"type:varchar(50);not null" json:"password" validate:"min=6,max=32"`
	FollowCount   int64  `gorm:"type:bigint;unsigned;not null;default:0" json:"follow_count"`
	FollowerCount int64  `gorm:"type:bigint;unsigned;not null;default:0" json:"follower_count"`
}

// Comment 评论：数据库实体
type Comment struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UserID    int64  `gorm:"type:BIGINT;not null;index:idx_user_id;评论用户ID" json:"user_id"`
	VideoID   int64  `gorm:"type:BIGINT;not null;index:idx_video_id;comment:被评论视频ID" json:"video_id"`
	Content   string `gorm:"type:varchar(300);not null;comment:评论内容" json:"content"`
}

// Favourite 点赞：数据库实体
type Favourite struct {
	ID      uint  `gorm:"primarykey"`
	UserID  int64 `gorm:"type:BIGINT;not null;uniqueIndex:idx_member_id;comment:点赞用户ID" json:"user_id"`
	VideoID int64 `gorm:"type:BIGINT;not null;uniqueIndex:idx_member_id;comment:被点赞视频ID" json:"video_id"`
	IsFavor int8  `gorm:"type:TINYINT;not null;comment:软删除的点赞记录" json:"is_favor"`
}

// Follow 关注：数据库实体
type Follow struct {
	ID         uint  `gorm:"primarykey"`
	FromUserID int64 `gorm:"type:BIGINT;not null;uniqueIndex:idx_member_id;comment:粉丝用户ID"`
	ToUserID   int64 `gorm:"type:BIGINT;not null;uniqueIndex:idx_member_id;comment:被关注用户ID"`
	IsFollow   int8  `gorm:"type:TINYINT;not null;comment:软删除的关注记录"`
}

// Message 消息：数据库实体
type Message struct {
	gorm.Model
	MessageID  int64  `gorm:"type:bigint;unsigned;not null;unique;uniqueIndex:idx_message_id" json:"message_id"`
	FromUserID int64  `gorm:"type:BIGINT;not null;index:idx_user_id;comment:发送用户ID"`
	ToUserId   int64  `gorm:"type:BIGINT;not null;index:idx_to_user_id;comment:接收用户ID"`
	Content    string `gorm:"type:varchar(300);not null;comment:聊天内容" json:"content"`
}
