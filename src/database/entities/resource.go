package entities

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

	// ReadNodes represents keys and roles which may be allowed or denied access to read.
	ReadNodes map[string]bool `json:"read_nodes"`
	// WriteNodes represents keys and roles which may be allowed or denied access to write.
	WriteNodes map[string]bool `json:"write_nodes"`
	// WriteNodes represents keys and roles which may be allowed or denied access to modify.
	ModifyNodes map[string]bool `json:"modify_nodes"`
	// DeleteNodes represents keys and roles which may be allowed or denied access to delete.
	DeleteNodes map[string]bool `json:"delete_nodes"`
}

func (r *Resource) PublicCanRead() bool {
	return (r.Flags & FlagPublicRead) == FlagPublicRead
}

// CheckRead checks if a given key (may be nil) has access to read from this resource.
func (r *Resource) CheckRead(key *Key) AccessStatus {
	if r.PublicCanRead() {
		return AccessAllowed
	}
	return checkRoleAccess(key, r.ReadNodes)
}

// CheckWrite checks if a given key has access to write to this resource.
func (r *Resource) CheckWrite(key *Key) AccessStatus {
	return checkRoleAccess(key, r.WriteNodes)
}

// CheckModify checks if a given key has access to modify this resource.
func (r *Resource) CheckModify(key *Key) AccessStatus {
	return checkRoleAccess(key, r.ModifyNodes)
}

// CheckDelete checks if a given key has access to delete this resource.
func (r *Resource) CheckDelete(key *Key) AccessStatus {
	return checkRoleAccess(key, r.DeleteNodes)
}

// checkRoleAccess returns the access status for a particular role in an access map.
func checkRoleAccess(key *Key, accessMap map[string]bool) AccessStatus {
	// if catch-all role is present, allow or deny based on its status.
	if status, ok := accessMap["*"]; ok {
		if status {
			return AccessAllowed
		}
		return AccessDenied
	}
	// check sorted key roles, returning allow or deny based on the first present in the access map.
	for _, role := range key.Roles {
		if status, ok := accessMap[role]; ok {
			if status {
				return AccessAllowed
			}
			return AccessDenied
		}
	}
	// no access specifiers on this level
	return AccessNeutral
}
