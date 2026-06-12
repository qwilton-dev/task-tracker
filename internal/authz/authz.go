package authz

type Role string

const (
	RoleOwner  Role = "owner"
	RoleMember Role = "member"
	RoleViewer Role = "viewer"
)

var roleHierarchy = map[Role]int{
	RoleOwner:  3,
	RoleMember: 2,
	RoleViewer: 1,
}

func (r Role) String() string {
	return string(r)
}

func (r Role) AtLeast(min Role) bool {
	return roleHierarchy[r] >= roleHierarchy[min]
}

type Action string

const (
	ActionCreateIssue       Action = "create_issue"
	ActionEditIssue         Action = "edit_issue"
	ActionDeleteIssue       Action = "delete_issue"
	ActionCreateProject     Action = "create_project"
	ActionEditProject       Action = "edit_project"
	ActionDeleteProject     Action = "delete_project"
	ActionManageLabels      Action = "manage_labels"
	ActionCreateInvite      Action = "create_invite"
	ActionChangeMemberRoles Action = "change_member_roles"
	ActionDeleteWorkspace   Action = "delete_workspace"
)

var actionMatrix = map[Action]Role{
	ActionCreateIssue:       RoleMember,
	ActionEditIssue:         RoleMember,
	ActionDeleteIssue:       RoleMember,
	ActionCreateProject:     RoleMember,
	ActionEditProject:       RoleMember,
	ActionDeleteProject:     RoleOwner,
	ActionManageLabels:      RoleMember,
	ActionCreateInvite:      RoleMember,
	ActionChangeMemberRoles: RoleOwner,
	ActionDeleteWorkspace:   RoleOwner,
}

func Can(userRole Role, action Action) bool {
	minRole, ok := actionMatrix[action]
	if !ok {
		return false
	}
	return userRole.AtLeast(minRole)
}
