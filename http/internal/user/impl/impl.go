package impl

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/sunflower10086/TikTok/http/internal/dao"
	"github.com/sunflower10086/TikTok/http/internal/models"
	"github.com/sunflower10086/TikTok/http/internal/pkg/result"
	"github.com/sunflower10086/TikTok/http/internal/user"
	"github.com/sunflower10086/TikTok/http/pkg/crypto"
	"github.com/sunflower10086/TikTok/http/pkg/jwt"
)

func Login(ctx context.Context, request *user.LoginRequest) (*user.LoginResponse, error) {
	// 检测用户是否存在
	userByUsername, err := dao.GetUserByUsername(request.Username)

	if err != nil {
		return nil, err
	}

	if userByUsername == nil {
		return nil, errors.New("账号密码错误")
	}

	if userByUsername.Password != crypto.SHA512Hash(request.Password) {
		return nil, errors.New("账号密码错误")
	}

	token, err := jwt.GenToken(int64(userByUsername.ID), userByUsername.Username)

	if err != nil {
		return nil, fmt.Errorf("生成token错误：%w", err)
	}

	userId := int64(userByUsername.ID)

	return &user.LoginResponse{
		Token:  &token,
		UserID: &userId,
	}, nil
}

func Register(ctx context.Context, request *user.RegisterRequest) (*user.RegisterResponse, error) {
	// 新用户注册时提供用户名，密码即可，
	// 用户名需要保证唯一。创建成功后返回用户 id 和权限token.

	// 检测用户名是否已经存在
	userByUsername, err := dao.GetUserByUsername(request.Username)

	if err != nil {
		return nil, err
	}

	if userByUsername != nil {
		return nil, errors.New("用户名已存在")
	}

	newUser := models.User{
		Username: request.Username,
		Password: crypto.SHA512Hash(request.Password), // MD5哈希加密
	}

	err = dao.CreateUser(&newUser)

	if err != nil {
		return nil, fmt.Errorf("创建用户错误：%w", err)
	}

	token, err := jwt.GenToken(newUser.ID, newUser.Username)

	if err != nil {
		return nil, fmt.Errorf("生成token错误：%w", err)
	}

	userId := newUser.ID

	return &user.RegisterResponse{
		Token:  &token,
		UserID: &userId,
	}, nil
}

func GetInfo(ctx *gin.Context, request *user.GetInfoRequest) (*user.GetInfoResponse, error) {
	userID := request.UserID

	userInfo, err := dao.GetUserByID(userID)

	if err != nil {
		log.Println(err.Error())
		return nil, err
	}

	_user := user.User{
		Avatar:          userInfo.Avatar,
		BackgroundImage: userInfo.BackgroundImage,
		FavoriteCount:   userInfo.OtherInfo.FavoriteCount,
		FollowCount:     userInfo.OtherInfo.FollowCount,
		FollowerCount:   userInfo.OtherInfo.FollowerCount,
		ID:              userInfo.ID,
		IsFollow:        dao.CheckIsFollowUser(ctx, int64(userID)),
		Name:            userInfo.Username,
		Signature:       userInfo.Signature,
		TotalFavorited:  userInfo.OtherInfo.TotalFavorited,
		WorkCount:       userInfo.OtherInfo.WorkCount,
	}

	return &user.GetInfoResponse{
		User:     &_user,
		Response: result.Response{StatusCode: result.SuccessCode},
	}, nil

}
