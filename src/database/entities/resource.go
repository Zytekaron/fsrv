package entities

import "fsrv/src/types"

type AccessStatus int8

const (
	AccessDenied AccessStatus = iota - 1
	AccessNeutral
	AccessAllowed
)

type Flags uint8

const (
	// FlagPublicRead means any user can read this file without a key.
	FlagPublicRead Flags = 1 << iota
)

// Resource represents an access specifier for a file or directory.
type Resource struct {
	// ID is the id of this Resource
	ID string `json:"id"`
	// Flags represents the access flags for this resource.
	Flags Flags `json:"flags"`

	// OperationNodes represents keys and roles which may be allowed or denied permission to perform operations.
	OperationNodes map[ResourceOperationAccess]bool `json:"nodes"`
}

type ResourceOperationAccess struct {
	// ID is the key id or role name for this node.
	ID string
	// Type is the operation type this node is for.
	Type types.OperationType
}

func (r *Resource) PublicCanRead() bool {
	return (r.Flags & FlagPublicRead) == FlagPublicRead
}

// CheckAccess checks if a given key (may be nil) has access to perform a particular operation on this resource.
func (r *Resource) CheckAccess(key *Key, op types.OperationType) AccessStatus {
	if key == nil {
		// allow public reads
		if op == types.OperationRead && r.PublicCanRead() {
			return AccessAllowed
		}
		return AccessDenied
	}

	return r.checkKeyAccess(key, op)
}

// checkKeyAccess returns the access status for a particular role in an access map.
func (r *Resource) checkKeyAccess(key *Key, op types.OperationType) AccessStatus {
	roa := ResourceOperationAccess{"*", op}
	// if catch-all role is present, allow or deny based on its status.
	if status, ok := r.OperationNodes[roa]; ok {
		if status {
			return AccessAllowed
		}
		return AccessDenied
	}

	// if the key itself is present, allow or deny based on its status.
	roa.ID = key.ID
	if status, ok := r.OperationNodes[roa]; ok {
		if status {
			return AccessAllowed
		}
		return AccessDenied
	}

	// check sorted key roles, returning allow or deny based on the first present in the access map.
	for _, role := range key.Roles {
		roa.ID = role
		if status, ok := r.OperationNodes[roa]; ok {
			if status {
				return AccessAllowed
			}
			return AccessDenied
		}
	}

	// no access specifiers on this level
	return AccessNeutral
}

func (r *Resource) GetID() string {
	return r.ID
}
