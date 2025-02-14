package repo

import (
	"context"
	"fmt"
	"os/user"
	"web-service/base"
	"web-service/base/apierrs"
	"web-service/model"

	"gorm.io/gorm"
)

type UserRepo struct {
	repo *gorm.DB
}

func NewUserRepo(repo *gorm.DB) *UserRepo {
	return &UserRepo{
		repo: repo,
	}
}

func (u *UserRepo) Create(ctx context.Context, user *model.User) (err error) {
	if user == nil {
		return apierrs.NewCreateError(fmt.Errorf("failed to create user, user is nil"))
	}
	err = u.repo.WithContext(ctx).Create(&user).Error
	if err != nil {
		return apierrs.NewCreateError(fmt.Errorf("failed to create user, %w", err))
	}
	return nil
}

func (u *UserRepo) Delete(ctx context.Context, user *model.User) (err error) {
	err = u.repo.WithContext(ctx).Delete(&user).Error
	if err != nil {
		return apierrs.NewDeleteError(fmt.Errorf("failed to delete user, %w", err))
	}
	return nil
}

func (u *UserRepo) Save(ctx context.Context, user *model.User) (err error) {
	if user == nil {
		return apierrs.NewCreateError(fmt.Errorf("failed to save user, user is nil"))
	}
	if err = u.repo.WithContext(ctx).Save(&user).Error; err != nil {
		return apierrs.NewSaveError(fmt.Errorf("failed to save user, %w", err))
	}
	return nil
}

func (u *UserRepo) GetUserByID(ctx context.Context, id uint, options ...base.UserQueryOption) (user *model.User, err error) {
	query := u.repo.WithContext(ctx).Model(&user).Where("id = ?", id)
	for _, option := range options {
		option(query)
	}

	if err = query.Take(&user).Error; err != nil {
		return nil, apierrs.NewGetError(fmt.Errorf("failed to get user, %w", err))
	}
	return user, nil
}

func (u *UserRepo) GetUserByName(ctx context.Context, name string, options ...base.UserQueryOption) (user *model.User, err error) {
	query := u.repo.WithContext(ctx).Model(&user).Where("name = ?", name)
	for _, option := range options {
		option(query)
	}
	if err = query.Take(&user).Error; err != nil {
		return nil, apierrs.NewGetError(fmt.Errorf("failed to get user, %w", err))
	}
	return user, nil
}

func (u *UserRepo) GetUserByEmail(ctx context.Context, email string, options ...base.UserQueryOption) (user *model.User, err error) {
	query := u.repo.WithContext(ctx).Model(&user).Where("email = ?", email)
	for _, option := range options {
		option(query)
	}
	if err = query.Take(&user).Error; err != nil {
		return nil, apierrs.NewGetError(fmt.Errorf("failed to get user, %w", err))
	}
	return user, nil
}

func (u *UserRepo) List(ctx context.Context, page, pageSize int) (total int64, users []*model.User, err error) {
	// 计数查询
	query := u.repo.WithContext(ctx).Model(&user.User{})
	if err = query.Count(&total).Error; err != nil {
		return 0, nil, apierrs.NewListError(fmt.Errorf("failed to count users, %w", err))

	}

	// 数据查询
	if err = query.
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&users).Error; err != nil {
		return 0, nil, apierrs.NewListError(fmt.Errorf("failed to list users, %w", err))
	}
	return total, users, nil
}

type GetUserRepo struct {
	*UserRepo
}

func NewGetUserRepo(userRepo *UserRepo) *GetUserRepo {
	return &GetUserRepo{
		UserRepo: userRepo,
	}
}
