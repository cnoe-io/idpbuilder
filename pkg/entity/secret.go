package entity

type Secret struct {
	IsCore    bool
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Username  string            `json:"username,omitempty"`
	Password  string            `json:"password,omitempty"`
	Token     string            `json:"token,omitempty"`
	Data      map[string]string `json:"data,omitempty"`
}
