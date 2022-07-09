package types

type AccessStatus int8

const (
	AccessAllowed AccessStatus = iota - 1
	AccessNeutral
	AccessDenied
)

// Permission represents an access specifier for a file or directory.
type Permission struct {
	// Public represents whether this permission should allow all access attempts.
	Public bool `json:"public"`
	// ReadKeys represents tokens which may be allowed or denied access to read.
	ReadKeys map[string]AccessStatus `json:"read_keys"`
	// ReadRoles represents roles which may be allowed or denied access to read.
	ReadRoles map[string]AccessStatus `json:"read_roles"`
	// WriteKeys represents tokens which may be allowed or denied access to write.
	WriteKeys map[string]AccessStatus `json:"write_keys"`
	// WriteRoles represents roles which may be allowed or denied access to write.
	WriteRoles map[string]AccessStatus `json:"write_roles"`
}

func (p *Permission) CheckRead(key *Key) AccessStatus {
	if p.Public {
		return AccessAllowed
	}
	if status, ok := p.ReadKeys[key.ID]; ok {
		return status
	}
	for _, role := range key.Roles {
		if status, ok := p.ReadRoles[role]; ok && status != AccessNeutral {
			return status
		}
	}
	return AccessNeutral
}
