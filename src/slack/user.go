package slack

import (
	"fmt"
)

// User holds information of users in a workspace
type User struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	RealName string `json:"real_name"`
	IsBot    bool   `json:"is_bot"`
}

// Members is a slice of User
type Members []User

// ID2UserName searches a user with the given user id and returns his/her user name
func (m Members) ID2UserName(id string) (string, error) {
	for _, u := range m {
		if u.ID == id {
			return u.Name, nil
		}
	}
	return "", fmt.Errorf("use id %s is not a member", id)
}
