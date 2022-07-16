package entities

import "fsrv/src/types"

type RolePerm struct {
	Role   string
	Status bool                //deny, allow
	TypeRW types.OperationType //read, write
}
