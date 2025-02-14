package helpers

import "web-service/model"

func GetCasbinRole(roleName string, policys []*model.Policy) [][]string {
	save := make([][]string, len(policys))
	for i, policy := range policys {
		save[i] = []string{roleName, policy.Path, policy.Method}
	}
	return save
}
