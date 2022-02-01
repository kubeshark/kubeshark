package workspace

type WorkspaceCreateRequest struct {
	Name       string   `json:"name" binding:"required"`
	Namespaces []string `json:"namespaces" binding:"required"`
}

type WorkspaceListItemResponse struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

type WorkspaceResponse struct {
	Id         string   `json:"id"`
	Name       string   `json:"name"`
	Namespaces []string `json:"namespaces"`
}
