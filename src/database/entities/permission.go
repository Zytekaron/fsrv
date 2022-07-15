package entities

type AccessStatus int8

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

// Permission represents an access specifier for a file or directory.
type Permission struct {
	// ID is the id of this permission
	ID int `json:"id"`
	// Public represents whether this permission should allow all access attempts.
	Public bool `json:"public"`

	// ReadNodes represents keys and roles which may be allowed or denied access to read.
	ReadNodes map[AccessNode]AccessStatus `json:"read_nodes"`
	// WriteNodes represents keys and roles which may be allowed or denied access to write.
	WriteNodes map[AccessNode]AccessStatus `json:"write_nodes"`
}

func (p *Permission) CheckRead(key *Key) AccessStatus {
	if p.Public {
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
