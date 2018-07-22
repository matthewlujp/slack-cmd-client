package slack

// Workspace holds basic information of a slack team
type Workspace struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Domain string `json:"domain"`
	Token  string `json:"-"`
}
