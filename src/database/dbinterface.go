package database

import (
	"errors"
	"fsrv/src/database/entities"
	"fsrv/src/types"
)

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

type DBInterface interface {
	Create(databaseFile string) (DBInterface, error) //creates the database if it does not exist
	Open(databaseFile string) (DBInterface, error)   //opens an existing database
	Exists(databaseFile string) error                //checks if the database exists
	Check() error                                    //checks database integrity
	Destroy() error                                  //destroys database objects but leaves database file intact

	CreateKey(key *entities.Key) error
	CreateResource(resource *entities.Resource) error
	CreateRole(role *entities.Role) error
	CreateRateLimit(keyid string, limit *entities.RateLimit)

	GetKeys() []*entities.Key
	GetKeyIDs() []string
	GetKeyData(keyid string) (*entities.Key, error)
	GetResources() []*entities.Resource
	GetResourceIDs() []string
	GetResourceData(resourceid string) (*entities.Resource, error)
	GetRoles() []string //

	GiveRole(keyid string, role ...string) error
	TakeRole(keyid string, role ...string) error
	GrantPermission(resource string, operationType types.OperationType, role ...string) []error
	RevokePermission(resource string, operationType types.OperationType, role ...string) []error
	SetRateLimit(keyid string, limit *entities.RateLimit)

	DeleteRole(name string) error
	DeleteKey(id string) error
	DeleteResource(id string) error
}
