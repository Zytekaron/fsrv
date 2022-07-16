package entities

type AccessStatus int8 // todo: consider removing this type in favor of using presence to represent neutrality

const (
	AccessDenied AccessStatus = iota - 1
	AccessNeutral
	AccessAllowed
)

type AccessNodeType int8

const (
	AccessNodeKey AccessNodeType = iota
	AccessNodeRole
)

type AccessNode struct {
	ID   string
	Type AccessNodeType
}

type Flags int8

const (
	// FlagPublicRead means any user can read this file without authentication.
	FlagPublicRead Flags = 1 << iota
	// FlagPublicWrite means any user can read this file without authentication.
	FlagPublicWrite
	// FlagAuthedRead means any authenticated can read this file.
	FlagAuthedRead
	// FlagAuthedWrite means any authenticated can read this file.
	FlagAuthedWrite
)

// Permission represents an access specifier for a file or directory.
type Permission struct {
	// ID is the id of this permission
	ID int `json:"id"`
	// Flags represents whether this permission should allow all access attempts.
	Flags Flags `json:"public"`

	// ReadNodes represents keys and roles which may be allowed or denied access to read.
	ReadNodes map[AccessNode]AccessStatus `json:"read_nodes"`
	// WriteNodes represents keys and roles which may be allowed or denied access to write.
	WriteNodes map[AccessNode]AccessStatus `json:"write_nodes"`
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

	if status, ok := p.ReadNodes[AccessNode{key.ID, AccessNodeKey}]; ok {
		return status
	}
	for _, role := range key.Roles {
		if status, ok := p.ReadNodes[AccessNode{role, AccessNodeRole}]; ok && status != AccessNeutral {
			return status
		}
	}
	return AccessNeutral
}
