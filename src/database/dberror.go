package database

import "errors"

var (
	ErrCreateFailed  = errors.New("database creation failed")
	ErrCheckFailed   = errors.New("database integrity check failed")
	ErrDestroyFailed = errors.New("database object wipe failed")

	ErrKeyDuplicate      = errors.New("A key with the given ID already exists")
	ErrRoleDuplicate     = errors.New("A role with the given ID already exists")
	ErrResourceDuplicate = errors.New("A resource with the given ID already exists")

	ErrKeyMissing      = errors.New("the specified key does not exist")
	ErrRoleMissing     = errors.New("the specified role does not exist")
	ErrResourceMissing = errors.New("the specified resource does not exist")

	ErrRoleNameBad     = errors.New("the given role name is not allowed")
	ErrKeyNameBad      = errors.New("the given key name is not allowed")
	ErrResourceNameBad = errors.New("the given resource name is not allowed")
)
