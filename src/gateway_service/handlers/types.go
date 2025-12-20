package handlers

type GetUserProjectsResponseItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AppsCount int32  `json:"apps_count"`
	CreatedAt string `json:"created_at"`
}

type GetUserProjectsResponse []GetUserProjectsResponseItem
