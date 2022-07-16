package database

import "fsrv/src/database/entities"

type DBInterface interface {
	Create() error  //creates the database if it does not exist
	Check() error   //checks database integrity
	Destroy() error //destroys database objects but leaves database file intact

	GetKeyRateLimit(keyid string) (entities.RateLimit, error)                                 //retrieves a key's RateLimit
	GetKeyRoles(keyid string) (map[string]struct{}, error)                                    //retrieves a key's Roles
	GetResourceRolePerms(resourceid string) ([]entities.RolePerm, error)                      //retrieves an entities RolePerms sorted by priority
	GetDirectKeyPerms(keyid string) ([]entities.RolePerm, error)                              //retrieves a key's specific permissions
	GetCombinedKeyResourcePerms(resourceid string, keyid string) ([]entities.RolePerm, error) //retrieves key specific and role specific permissions sorted by priority

	GetKeys() []*entities.Key
	GetRoles() []*string
	GetResources() []*entities.Resource
	GetKeyData() *entities.Key
	GetRoleData() *string

	CreateRole(name string, precedence int)
	DeleteRole(name string)
	GrantPermission(resource string, role string)
}
