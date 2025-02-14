package service

import (
	"context"
	"errors"
	"fmt"
	"web-service/base"
	"web-service/base/apierrs"
	"web-service/base/helpers"
	"web-service/model"
	"web-service/pkg/permissions"
	"web-service/schema"
)

type RoleService struct {
	generalRoleRepo  base.GeneralRoleRepoer
	getPolicyRepo    base.GetPolicyRepoer
	appendPolicyRepo base.AssociationPolicyer
	casbinRepo       permissions.GeneralAuthorizer
}

func NewRoleService(
	repo base.GeneralRoleRepoer,
	policyRepo base.GetPolicyRepoer,
	appendRepo base.AssociationPolicyer,
	casbinRepo permissions.GeneralAuthorizer,
) *RoleService {
	return &RoleService{
		generalRoleRepo:  repo,
		getPolicyRepo:    policyRepo,
		casbinRepo:       casbinRepo,
		appendPolicyRepo: appendRepo,
	}
}

func (r *RoleService) CreateRole(ctx context.Context, req *schema.RoleCreateRequest) (err error) {
	role := &model.Role{
		Name:        req.Name,
		Description: req.Desc,
	}
	return r.generalRoleRepo.Create(ctx, role)
}

// DeleteRole 删除角色
func (r *RoleService) DeleteRole(ctx context.Context, req *schema.IDRequest) (err error) {
	role, err := r.generalRoleRepo.GetRoleByID(ctx, req.ID, base.WithRoleUsers())
	if err != nil {
		return err
	}
	if role.Users != nil && len(role.Users) > 0 {
		var userNames []string
		for _, user := range role.Users {
			userNames = append(userNames, user.Name)
		}
		return apierrs.NewDeleteError(fmt.Errorf("failed to delete role, role has users: %s", userNames))
	}
	return r.generalRoleRepo.Delete(ctx, role)
}

// UpdateRoleDesc 更新角色描述信息
func (r *RoleService) UpdateRoleDesc(ctx context.Context, req *schema.RoleUpdateRequest) (err error) {
	role, err := r.generalRoleRepo.GetRoleByID(ctx, req.ID)
	if err != nil {
		return err
	}

	if role.Description == req.Desc {
		return nil
	}

	role.Description = req.Desc
	return r.generalRoleRepo.Save(ctx, role)
}

// AddByPolicy 增加 casbin 角色权限
func (r *RoleService) AddByPolicy(ctx context.Context, req *schema.RoleUpdatePolicyRequest) (err error) {
	role, err := r.generalRoleRepo.GetRoleByID(ctx, req.RoleID, base.WithRolePolicys())
	if err != nil {
		return err
	}
	reqPolicys, err := r.getPolicyRepo.GetPolicyByIDs(ctx, req.PolicyID)
	if err != nil {
		return err
	}
	if len(reqPolicys) != len(req.PolicyID) {
		return apierrs.NewUpdateError(errors.New("failed to update policy, policy not exists"))
	}

	err = r.appendPolicyRepo.AppendPolicy(ctx, role, reqPolicys)
	if err != nil {
		return err
	}

	// 更新 casbin 策略
	save := helpers.GetCasbinRole(role.Name, reqPolicys)
	return r.casbinRepo.CreateRolePolicys(ctx, save)
}

// DeleteByPolicy 删除 casbin 角色权限
func (r *RoleService) DeleteByPolicy(ctx context.Context, req *schema.RoleDeltePolicyRequest) (err error) {
	role, err := r.generalRoleRepo.GetRoleByID(ctx, req.RoleID, base.WithRolePolicys())
	if err != nil {
		return err
	}
	reqPolicys, err := r.getPolicyRepo.GetPolicyByIDs(ctx, req.PolicyID)
	if err != nil {
		return err
	}

	if len(reqPolicys) != len(req.PolicyID) {
		return apierrs.NewUpdateError(errors.New("failed to update policy, policy not exists"))
	}

	if err := r.appendPolicyRepo.DeletePolicy(ctx, role, reqPolicys); err != nil {
		return err
	}

	delete := helpers.GetCasbinRole(role.Name, reqPolicys)
	return r.casbinRepo.DeleteRolePolicys(ctx, delete)
}

func (r *RoleService) GetRoleByID(ctx context.Context, req *schema.IDRequest) (role *model.Role, err error) {
	return r.generalRoleRepo.GetRoleByID(ctx, req.ID)
}

func (r *RoleService) ListRole(ctx context.Context, req *schema.RoleListRequest) (data *schema.RoleListResponse, err error) {
	total, roles, err := r.generalRoleRepo.List(ctx, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	return &schema.RoleListResponse{
		Total: total,
		ListRequest: &schema.ListRequest{
			Page:     req.Page,
			PageSize: req.PageSize,
		},
		Items: roles,
	}, nil
}
