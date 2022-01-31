package user

import "mizuserver/pkg/providers/workspace"

type InviteStatus string

const (
	PendingInviteStatus  InviteStatus = "pending"
	AcceptedInviteStatus InviteStatus = "active"
)

type TokenResponse struct {
	Token string `json:"token"`
}

type User struct {
	Username    string       `json:"username"`
	UserId      string       `json:"userId"`
	Status      InviteStatus `json:"status"`
	WorkspaceId string       `json:"workspaceId"`
	SystemRole  string       `json:"role"`
}

type UserListItem struct {
	Username   string       `json:"username"`
	UserId     string       `json:"userId"`
	Status     InviteStatus `json:"status"`
	SystemRole string       `json:"role"`
}

type InviteUserRequest struct {
	Username    string `json:"username"`
	WorkspaceId string `json:"workspaceId"`
	SystemRole  string `json:"role"`
}

type EditUserRequest struct {
	WorkspaceId string `json:"workspaceId"`
	SystemRole  string `json:"role"`
}

type WhoAmIResponse struct {
	Username   string                       `json:"username"`
	SystemRole string                       `json:"role"`
	Workspace  *workspace.WorkspaceResponse `json:"workspace"`
}
