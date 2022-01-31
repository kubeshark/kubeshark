package workspace

type WorkspaceCreateRequest struct {
	Name       string   `json:"name"`
	Namespaces []string `json:"namespaces"`
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
