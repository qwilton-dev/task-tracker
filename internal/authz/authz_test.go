package authz

import "testing"

func TestRole_AtLeast(t *testing.T) {
	tests := []struct {
		user Role
		min  Role
		want bool
	}{
		{RoleOwner, RoleOwner, true},
		{RoleOwner, RoleMember, true},
		{RoleOwner, RoleViewer, true},
		{RoleMember, RoleOwner, false},
		{RoleMember, RoleMember, true},
		{RoleMember, RoleViewer, true},
		{RoleViewer, RoleOwner, false},
		{RoleViewer, RoleMember, false},
		{RoleViewer, RoleViewer, true},
	}
	for _, tt := range tests {
		if got := tt.user.AtLeast(tt.min); got != tt.want {
			t.Errorf("%s.AtLeast(%s) = %v, want %v", tt.user, tt.min, got, tt.want)
		}
	}
}

func TestCan_Matrix(t *testing.T) {
	tests := []struct {
		role   Role
		action Action
		want   bool
	}{
		{RoleOwner, ActionCreateIssue, true},
		{RoleOwner, ActionDeleteProject, true},
		{RoleOwner, ActionChangeMemberRoles, true},
		{RoleOwner, ActionDeleteWorkspace, true},

		{RoleMember, ActionCreateIssue, true},
		{RoleMember, ActionEditIssue, true},
		{RoleMember, ActionDeleteIssue, true},
		{RoleMember, ActionCreateProject, true},
		{RoleMember, ActionEditProject, true},
		{RoleMember, ActionManageLabels, true},
		{RoleMember, ActionCreateInvite, true},
		{RoleMember, ActionDeleteProject, false},
		{RoleMember, ActionChangeMemberRoles, false},
		{RoleMember, ActionDeleteWorkspace, false},

		{RoleViewer, ActionCreateIssue, false},
		{RoleViewer, ActionEditIssue, false},
		{RoleViewer, ActionDeleteIssue, false},
		{RoleViewer, ActionCreateProject, false},
		{RoleViewer, ActionDeleteProject, false},
		{RoleViewer, ActionManageLabels, false},
		{RoleViewer, ActionChangeMemberRoles, false},
		{RoleViewer, ActionDeleteWorkspace, false},
	}
	for _, tt := range tests {
		if got := Can(tt.role, tt.action); got != tt.want {
			t.Errorf("Can(%s, %s) = %v, want %v", tt.role, tt.action, got, tt.want)
		}
	}
}

func TestCan_UnknownAction(t *testing.T) {
	if Can(RoleOwner, "nonexistent_action") {
		t.Error("expected false for unknown action")
	}
}
