package entities

type PermissionUpdates struct {
	public       *bool
	addReadKeys  *map[string]AccessStatus
	delReadKeys  *map[string]AccessStatus
	addWriteKeys *map[string]AccessStatus
	delWriteKeys *map[string]AccessStatus
}

func NewPermissionUpdates() *PermissionUpdates {
	return &PermissionUpdates{}
}

func (u *PermissionUpdates) WithPublic(public bool) *PermissionUpdates {
	*u.public = public
	return u
}

func (u *PermissionUpdates) AddReadKeys(keys map[string]AccessStatus) *PermissionUpdates {
	*u.addReadKeys = keys
	return u
}

func (u *PermissionUpdates) RemoveReadKeys(keys map[string]AccessStatus) *PermissionUpdates {
	*u.delReadKeys = keys
	return u
}

func (u *PermissionUpdates) AddWriteKeys(keys map[string]AccessStatus) *PermissionUpdates {
	*u.addWriteKeys = keys
	return u
}

func (u *PermissionUpdates) RemoveWriteKeys(keys map[string]AccessStatus) *PermissionUpdates {
	*u.delWriteKeys = keys
	return u
}
