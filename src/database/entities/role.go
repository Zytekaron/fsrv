package entities

type Role struct {
	Roleid   int
	RoleName string
}

type RolePerm struct {
	Role      Role
	AccessDAA int8 //deny, agnostic, allow
	TypeRW    bool //read, write
}
