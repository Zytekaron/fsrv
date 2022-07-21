package entities

import "fsrv/src/types"

type Role struct {
	ID         string
	Precedence int
}

type RolePerm struct {
	Role   Role
	Status bool                //deny, allow
	TypeRW types.OperationType //read, write
}
