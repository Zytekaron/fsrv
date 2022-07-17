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

type Flags int8

const (
	// FlagPublicRead means any user can read this file without authentication.
	FlagPublicRead Flags = 1 << iota
	// FlagAuthedRead means any authenticated can read this file.
	FlagAuthedRead
	// FlagPublicWrite means any user can write this file without authentication.
	FlagPublicWrite
	// FlagAuthedWrite means any authenticated can read this file.
	FlagAuthedWrite
	// FlagPublicModify means any user can modify this file without authentication.
	FlagPublicModify
	// FlagAuthedModify means any authenticated can modify this file.
	FlagAuthedModify
	// FlagPublicDelete means any user can delete this file without authentication.
	FlagPublicDelete
	// FlagAuthedDelete means any authenticated can delete this file.
	FlagAuthedDelete
)

// Resource represents an access specifier for a file or directory.
type Resource struct {
	// ID is the id of this Resource
	ID int `json:"id"`
	// Flags represents whether this permission should allow all access attempts.
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

func GetFlags(authed bool, accessType types.OperationType) Flags {
	var f Flags
	switch accessType {
	case types.OperationRead:
		if authed {
			f = FlagAuthedRead
		} else {
			f = FlagPublicRead
		}
	case types.OperationWrite:
		if authed {
			f = FlagAuthedWrite
		} else {
			f = FlagPublicWrite
		}
	case types.OperationModify:
		if authed {
			f = FlagAuthedModify
		} else {
			f = FlagPublicModify
		}
	case types.OperationDelete:
		if authed {
			f = FlagAuthedDelete
		} else {
			f = FlagPublicDelete
		}
	}
	return f
}

func (p *Resource) CheckAccessFlags(authed bool, accessType types.OperationType) AccessStatus {
	flags := GetFlags(authed, accessType)

	if authed {
		if (p.Flags & flags) == flags {
			return AccessAllowed
		} else {
			return AccessNeutral
		}
	} else {
		if (p.Flags & flags) == flags {
			return AccessAllowed
		} else {
			return AccessDenied
		}
	}
}

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

func (p *Resource) CheckAccessGeneral(key *Key, accessMap map[string]bool, accessType types.OperationType) AccessStatus {
	switch p.CheckAccessFlags(key != nil, accessType) {
	case AccessAllowed:
		return AccessAllowed
	case AccessDenied:
		return AccessDenied
	case AccessNeutral:
		//do nothing and continue
	default:
		//todo: make into real error
		fmt.Println("[error] its all gone horribly wrong... (permission.go)")
	}

	return CheckRoleAccess(key, accessMap)
}

func (p *Resource) CheckRead(key *Key) AccessStatus {
	return p.CheckAccessGeneral(key, p.ReadNodes, types.OperationRead)
}

func (p *Resource) CheckWrite(key *Key) AccessStatus {
	return p.CheckAccessGeneral(key, p.WriteNodes, types.OperationRead)
}

func (p *Resource) CheckModify(key *Key) AccessStatus {
	return p.CheckAccessGeneral(key, p.ModifyNodes, types.OperationModify)
}

func (p *Resource) CheckDelete(key *Key) AccessStatus {
	return p.CheckAccessGeneral(key, p.DeleteNodes, types.OperationDelete)
}

func (p *Resource) CheckReadZytekaron(key *Key) AccessStatus {
	if key == nil {
		if (p.Flags & FlagPublicRead) == FlagPublicRead {
			return AccessAllowed
		}
		return AccessDenied
	}
	if (p.Flags & FlagAuthedRead) == FlagAuthedRead {
		return AccessAllowed
	}

	if status, ok := p.ReadNodes[key.ID]; ok {
		if status {
			return AccessAllowed
		}
		return AccessDenied
	}

	for _, role := range key.Roles {
		if status, ok := p.ReadNodes[role]; ok {
			if status {
				return AccessAllowed
			}
			return AccessDenied
		}
	}
	return AccessNeutral
}
