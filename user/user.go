package user

type User struct {
	ID       string `json:"_id"`
	Name     string `json:"name"`
	Password string `json:"password"`
}
