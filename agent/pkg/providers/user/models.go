package user

import "github.com/up9inc/mizu/agent/pkg/providers/workspace"

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
	Username   string                               `json:"username"`
	UserId     string                               `json:"userId"`
	Status     InviteStatus                         `json:"status"`
	SystemRole string                               `json:"role"`
	Workspace  *workspace.WorkspaceListItemResponse `json:"workspace"`
}

type InviteUserRequest struct {
	Username    string `json:"username" binding:"required"`
	WorkspaceId string `json:"workspaceId" binding:"required"`
	SystemRole  string `json:"role" binding:"required,eq=admin|eq=user"`
}

type EditUserRequest struct {
	WorkspaceId string `json:"workspaceId"`
	SystemRole  string `json:"role" binding:"required,eq=admin|eq=user"`
}

type WhoAmIResponse struct {
	Username   string                       `json:"username"`
	SystemRole string                       `json:"role"`
	Workspace  *workspace.WorkspaceResponse `json:"workspace"`
}

type ErrorUserNotFound struct{}

func (e *ErrorUserNotFound) Error() string {
	return "user not found"
}
