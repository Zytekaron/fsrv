package entities

import (
	"fmt"
	"fsrv/src/types"
)

type AccessStatus int8

const (
	AccessDenied AccessStatus = iota - 1
	AccessNeutral
	AccessAllowed
)

type Flags uint8

const (
	// FlagPublicRead means any user can read this file without auth.
	FlagPublicRead Flags = 1 << iota
	// FlagAuthedRead means any authed user can read this file.
	FlagAuthedRead
	// FlagPublicWrite means any user can write this file without auth.
	FlagPublicWrite
	// FlagAuthedWrite means any authed user can read this file.
	FlagAuthedWrite
	// FlagPublicModify means any user can modify this file without auth.
	FlagPublicModify
	// FlagAuthedModify means any authed user can modify this file.
	FlagAuthedModify
	// FlagPublicDelete means any user can delete this file without auth.
	FlagPublicDelete
	// FlagAuthedDelete means any authed user can delete this file.
	FlagAuthedDelete
)

// Resource represents an access specifier for a file or directory.
type Resource struct {
	// ID is the id of this Resource
	ID string `json:"id"`
	// Flags represents the access flags for this resource.
	Flags Flags `json:"flags"`

	// ReadNodes represents keys and roles which may be allowed or denied access to read.
	ReadNodes map[string]bool `json:"read_nodes"`
	// WriteNodes represents keys and roles which may be allowed or denied access to write.
	WriteNodes map[string]bool `json:"write_nodes"`
	// WriteNodes represents keys and roles which may be allowed or denied access to modify.
	ModifyNodes map[string]bool `json:"modify_nodes"`
	// DeleteNodes represents keys and roles which may be allowed or denied access to delete.
	DeleteNodes map[string]bool `json:"delete_nodes"`
}

func GetFlag(authed bool, accessType types.OperationType) Flags {
	switch accessType {
	case types.OperationRead:
		if authed {
			return FlagAuthedRead
		}
		return FlagPublicRead
	case types.OperationWrite:
		if authed {
			return FlagAuthedWrite
		}
		return FlagPublicWrite
	case types.OperationModify:
		if authed {
			return FlagAuthedModify
		}
		return FlagPublicModify
	case types.OperationDelete:
		if authed {
			return FlagAuthedDelete
		}
		return FlagPublicDelete
	}
	//todo: make into real error
	panic("it's all gone horribly wrong... (resource.go GetFlag)")
}

func (p *Resource) CheckAccessFlags(authed bool, accessType types.OperationType) AccessStatus {
	flag := GetFlag(authed, accessType)

	if (p.Flags & flag) == flag {
		return AccessAllowed
	}
	return AccessDenied
}

// CheckRoleAccess returns the access status for a particular role in an access map.
func CheckRoleAccess(key *Key, accessMap map[string]bool) AccessStatus {
	for _, role := range key.Roles {
		if status, ok := accessMap[role]; ok {
			if status {
				return AccessAllowed
			}
			return AccessDenied
		}
	}
	return AccessNeutral
}

// CheckAccess checks if a given key (may be nil) has access to this resource for a particular operation.
func (p *Resource) CheckAccess(key *Key, accessMap map[string]bool, accessType types.OperationType) AccessStatus {
	authed := key != nil
	switch p.CheckAccessFlags(authed, accessType) {
	case AccessAllowed:
		return AccessAllowed
	case AccessDenied:
		return AccessDenied
	case AccessNeutral:
		// do nothing and continue
	default:
		//todo: make into real error
		fmt.Println("[error] its all gone horribly wrong... (resource.go CheckAccess)")
	}
	// no further checks possible for unauthenticated users
	if !authed {
		return AccessDenied
	}

	return CheckRoleAccess(key, accessMap)
}

// CheckRead checks if a given key (may be nil) has access to read from this resource.
func (p *Resource) CheckRead(key *Key) AccessStatus {
	return p.CheckAccess(key, p.ReadNodes, types.OperationRead)
}

// CheckWrite checks if a given key (may be nil) has access to write to this resource.
func (p *Resource) CheckWrite(key *Key) AccessStatus {
	return p.CheckAccess(key, p.WriteNodes, types.OperationRead)
}

// CheckModify checks if a given key (may be nil) has access to modify this resource.
func (p *Resource) CheckModify(key *Key) AccessStatus {
	return p.CheckAccess(key, p.ModifyNodes, types.OperationModify)
}

// CheckDelete checks if a given key (may be nil) has access to delete this resource.
func (p *Resource) CheckDelete(key *Key) AccessStatus {
	return p.CheckAccess(key, p.DeleteNodes, types.OperationDelete)
}
