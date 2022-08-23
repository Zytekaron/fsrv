package entities

import "fsrv/src/types"

type Role struct {
	ID         string
	Precedence int
}

type Permission struct {
	ResourceID string
	TypeRWMD   types.OperationType //read, write, modify, delete
	Status     bool                //deny, allow
}

type RolePerm struct {
	Role Role
	Perm Permission
}

func (r *Role) GetID() string {
	return r.ID
}
