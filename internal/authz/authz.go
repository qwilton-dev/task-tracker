package authz

type Role string

const (
	RoleAdmin  Role = "admin"
	RoleMember Role = "member"
	RoleViewer Role = "viewer"
)

func (r Role) String() string {
	return string(r)
}
