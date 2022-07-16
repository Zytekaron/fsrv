package database

import "fsrv/src/types"

type PermissionCache struct{}

type PermissionController struct{}

type PermissionInterface interface {
	validateRequest(keyid string, resourceid string, requestType types.OperationType) error
}
