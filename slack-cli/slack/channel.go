package slack

type Channel struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Members         []string `json:"members"`
	IsMember        bool     `json:"is_member"`
	Purpose         Purpose  `json:"purpose"`
	IsDirectMessage bool     `json:"is_im"`
	User            string   `json:"user"`
}

type Purpose struct {
	Value string `json:"value"`
}
