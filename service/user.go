package service

import (
	"context"
	"errors"
	"fmt"
	"web-service/base"
	"web-service/base/apierrs"
	"web-service/base/constant"
	"web-service/base/data"
	"web-service/model"
	"web-service/pkg/jwt"
	"web-service/pkg/permissions"
	"web-service/schema"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	userRepo    base.GeneralUserRepoer
	getRoleRepo base.GetRoleRepoer
	cache       base.Cache
	casbin      permissions.GeneralAuthorizer
}

func NewUserService(userRepo base.GeneralUserRepoer, getRoleRepo base.GetRoleRepoer, cache base.Cache, casbin permissions.GeneralAuthorizer) *UserService {
	return &UserService{
		userRepo:    userRepo,
		getRoleRepo: getRoleRepo,
		cache:       cache,
		casbin:      casbin,
	}
}

func (u *UserService) RegistryUser(ctx context.Context, req *schema.UserRegistryRequest) (err error) {
	if _, err = u.userRepo.GetUserByName(ctx, req.Name); err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		encryptPassword, err := u.encryptPassword(ctx, req.Password)
		if err != nil {
			return err
		}
		err = u.userRepo.Create(ctx, &model.User{
			Name:     req.Name,
			Password: encryptPassword,
			Avatar:   req.Avatar,
			Email:    req.Email,
			Mobile:   req.Mobile,
		})
		if err != nil {
			return apierrs.NewCreateError(err)
		}
		return nil
	}
	return apierrs.NewCreateError(fmt.Errorf("user %s already exists", req.Name))
}

func (u *UserService) Login(ctx context.Context, req *schema.UserLoginRequest) (res *schema.UserLoginResponse, err error) {
	var user *model.User
	if req.Email != "" {
		user, err = u.userRepo.GetUserByEmail(ctx, req.Email, base.WithUserRole())
		if err != nil {
			return nil, err
		}
	} else {
		user, err = u.userRepo.GetUserByName(ctx, req.Username, base.WithUserRole())
		if err != nil {
			return nil, err
		}
	}

	if user.Status == model.StatusDisabled {
		return nil, apierrs.NewAuthError(fmt.Errorf("user %s not found", req.Username))
	}
	if !u.verifyPassword(ctx, req.Password, user.Password) {
		return nil, apierrs.NewAuthError(errors.New("invalid password"))
	}

	err = u.cache.SetString(ctx, constant.RoleCacheKeyPrefix+user.Name, user.Role.Name, &data.NeverExpires)
	if err != nil {
		return nil, err
	}

	token, err := jwt.NewClaims(user.ID, user.Name).GenerateToken()
	if err != nil {
		return nil, err
	}
	res = &schema.UserLoginResponse{
		User: &schema.UserResponse{
			MetaData: &model.MetaData{ID: user.ID, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt},
			Name:     user.Name,
			Avatar:   user.Avatar,
			Email:    user.Email,
			Mobile:   user.Mobile,
			Role:     user.Role,
			RoleID:   user.RoleID,
			Status:   user.Status,
		},
		Token: token,
	}
	return res, err
}

func (u *UserService) DeleteUser(ctx context.Context, req *schema.IDRequest) (err error) {
	var user *model.User
	user, err = u.userRepo.GetUserByID(ctx, req.ID)
	if err != nil {
		return err
	}
	if user.Status == model.StatusDisabled {
		return apierrs.NewUpdateError(fmt.Errorf("user %d not found", req.ID))
	}

	user.Status = model.StatusDisabled
	err = u.userRepo.Save(ctx, user)
	if err != nil {
		return err
	}
	return u.userRepo.Delete(ctx, user)
}

func (u *UserService) UpdatePassword(ctx context.Context, req *schema.UserUpdatePasswordRequest) (err error) {
	var user *model.User
	user, err = u.userRepo.GetUserByID(ctx, req.ID)
	if err != nil {
		return err
	}
	if user.Status == model.StatusDisabled {
		return apierrs.NewUpdateError(fmt.Errorf("user %d not found", req.ID))
	}

	if !u.verifyPassword(ctx, req.OldPassword, user.Password) {
		return apierrs.NewUpdateError(fmt.Errorf("invalid password"))
	}
	encryptPassword, err := u.encryptPassword(ctx, req.NewPassword)
	if err != nil {
		return err
	}
	user.Password = encryptPassword
	return u.userRepo.Save(ctx, user)
}

func (u *UserService) UpdateUser(ctx context.Context, req *schema.UserUpdateRequest) (err error) {
	var user *model.User
	user, err = u.userRepo.GetUserByID(ctx, req.ID)
	if err != nil {
		return err
	}
	if user.Status == model.StatusDisabled {
		return apierrs.NewUpdateError(fmt.Errorf("user %d not found", req.ID))
	}

	user.Avatar = req.Avatar
	user.Email = req.Email
	user.Mobile = req.Mobile
	return u.userRepo.Save(ctx, user)
}

// UpdateUserRole 更新用户角色
func (u *UserService) UpdateUserRole(ctx context.Context, req *schema.UserUpdateRoleRequest) (err error) {
	var user *model.User
	user, err = u.userRepo.GetUserByID(ctx, req.UserID)
	if err != nil {
		return err
	}
	if user.Status == model.StatusDisabled {
		return apierrs.NewUpdateError(fmt.Errorf("user %d not found", req.UserID))
	}

	if user.RoleID == req.RoleID {
		return apierrs.NewUpdateError(fmt.Errorf("user %s already in role %s", user.Name, user.Role.Name))
	}
	user.RoleID = req.RoleID
	role, err := u.getRoleRepo.GetRoleByID(ctx, req.RoleID)
	if err != nil {
		return err
	}
	if err := u.cache.Del(ctx, constant.RoleCacheKeyPrefix+user.Name); err != nil {
		return err
	}
	if err := u.userRepo.Save(ctx, user); err != nil {
		return err
	}
	if err := u.cache.Del(ctx, constant.RoleCacheKeyPrefix+user.Name); err != nil {
		return err
	}

	return u.cache.SetString(ctx, constant.RoleCacheKeyPrefix+user.Name, role.Name, &data.NeverExpires)
}

func (u *UserService) GetUserBasicInfoByID(ctx context.Context, req *schema.IDRequest) (res *schema.UserResponse, err error) {
	user, err := u.userRepo.GetUserByID(ctx, req.ID, base.WithUserRole(), base.WithUserPolicys())
	if err != nil {
		return nil, err
	}
	if user.Status == model.StatusDisabled {
		return nil, apierrs.NewGetError(fmt.Errorf("user %d not found", req.ID))
	}

	return &schema.UserResponse{
		MetaData: &model.MetaData{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
		Name:   user.Name,
		Avatar: user.Avatar,
		Email:  user.Email,
		Status: user.Status,
		Mobile: user.Mobile,
		Role:   user.Role,
		RoleID: user.RoleID,
	}, nil
}

func (u *UserService) ListUser(ctx context.Context, req *schema.UserListRequest) (data *schema.UserListResponse, err error) {
	total, users, err := u.userRepo.List(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	return u.formatUserList(req, total, users), nil
}

func (u *UserService) formatUserList(req *schema.UserListRequest, total int64, users []*model.User) *schema.UserListResponse {
	data := &schema.UserListResponse{
		ListRequest: &schema.ListRequest{
			Page:     req.Page,
			PageSize: req.PageSize,
		},
		Total: total,
		Items: make([]*schema.UserResponse, 0, len(users)),
	}

	for _, user := range users {
		meta := &model.MetaData{
			ID:        user.ID,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}

		data.Items = append(data.Items, &schema.UserResponse{
			MetaData: meta,
			Name:     user.Name,
			Avatar:   user.Avatar,
			Email:    user.Email,
			Mobile:   user.Mobile,
			Status:   user.Status,
			RoleID:   user.RoleID,
		})
	}

	return data
}

// encryptPassword 加密密码
func (us *UserService) encryptPassword(_ context.Context, Pass string) (string, error) {
	hashPwd, err := bcrypt.GenerateFromPassword([]byte(Pass), bcrypt.DefaultCost)
	if err != nil {
		return "", apierrs.NewAuthError(fmt.Errorf("failed to encrypt password, %w", err))
	}
	return string(hashPwd), nil
}

// verifyPassword 验证密码
func (us *UserService) verifyPassword(_ context.Context, loginPass, userPass string) bool {
	if len(loginPass) == 0 && len(userPass) == 0 {
		return true
	}
	err := bcrypt.CompareHashAndPassword([]byte(userPass), []byte(loginPass))
	return err == nil
}
