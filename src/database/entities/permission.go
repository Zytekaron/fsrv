package entities

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
)

// Permission represents an access specifier for a file or directory.
type Permission struct {
	// ID is the id of this permission
	ID int `json:"id"`
	// Flags represents whether this permission should allow all access attempts.
	Flags Flags `json:"flags"`

	// ReadNodes represents keys and roles which may be allowed or denied access to read.
	ReadNodes map[string]bool `json:"read_nodes"`
	// WriteNodes represents keys and roles which may be allowed or denied access to write.
	WriteNodes map[string]bool `json:"write_nodes"`
	// WriteNodes represents keys and roles which may be allowed or denied access to modify.
	ModifyNodes map[string]bool `json:"modify_nodes"`
}

func (p *Permission) CheckRead(key *Key) AccessStatus {
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
