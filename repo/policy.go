package repo

import (
	"context"
	"fmt"
	"web-service/base"
	"web-service/base/apierrs"
	"web-service/model"

	"gorm.io/gorm"
)

type PolicyRepo struct {
	data *gorm.DB
}

func NewPolicyRepo(repo *gorm.DB) *PolicyRepo {
	return &PolicyRepo{
		data: repo,
	}
}

func (p *PolicyRepo) Create(ctx context.Context, policy *model.Policy) (err error) {
	if err = p.data.WithContext(ctx).Create(&policy).Error; err != nil {
		return apierrs.NewCreateError(fmt.Errorf("failed to create policy, %w", err))
	}
	return nil
}

func (p *PolicyRepo) Save(ctx context.Context, policy *model.Policy) (err error) {
	if err = p.data.WithContext(ctx).Save(&policy).Error; err != nil {
		return apierrs.NewSaveError(fmt.Errorf("failed to save policy, %w", err))
	}
	return nil
}

func (p *PolicyRepo) Delete(ctx context.Context, policy *model.Policy) (err error) {
	if err = p.data.WithContext(ctx).Delete(policy).Error; err != nil {
		return apierrs.NewDeleteError(fmt.Errorf("failed to delete policy, %w", err))
	}
	return nil
}

func (p *PolicyRepo) List(ctx context.Context, page int, pageSize int) (total int64, policys []*model.Policy, err error) {
	query := p.data.WithContext(ctx).Model(&model.Policy{})
	err = query.Count(&total).Error
	if err != nil {
		return 0, nil, apierrs.NewListError(fmt.Errorf("failed to count policies, %w", err))
	}

	err = query.
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&policys).Error
	if err != nil {
		return 0, nil, apierrs.NewListError(fmt.Errorf("failed to list policies, %w", err))
	}
	return total, policys, nil
}

func (p *PolicyRepo) GetPolicyByID(ctx context.Context, id uint, options ...base.PolicyQueryOption) (policy *model.Policy, err error) {
	query := p.data.WithContext(ctx).Model(policy).Where("id = ?", id)
	for _, option := range options {
		option(query)
	}

	if err = query.Take(&policy).Error; err != nil {
		return nil, apierrs.NewGetError(fmt.Errorf("failed to get policy, %w", err))
	}
	return policy, nil
}

func (p *PolicyRepo) GetPolicyByName(ctx context.Context, name string, options ...base.PolicyQueryOption) (policy *model.Policy, err error) {
	query := p.data.WithContext(ctx).Model(policy).Where("name = ?", name)
	for _, option := range options {
		option(query)
	}

	if err = query.First(&policy).Error; err != nil {
		return nil, apierrs.NewGetError(fmt.Errorf("failed to get policy, %w", err))
	}
	return policy, nil
}

func (p *PolicyRepo) GetPolicyByIDs(ctx context.Context, ids []uint) (policys []*model.Policy, err error) {
	if err = p.data.WithContext(ctx).Where("id in (?)", ids).Find(&policys).Error; err != nil {
		return nil, apierrs.NewGetError(fmt.Errorf("failed to get policy, %w", err))
	}
	return policys, nil
}
