package user

type User struct {
	ID       string `json:"_id"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type UserP struct {
	ID       string `json:"Id"`
	Name     string `json:"Name"`
	Password string `json:"Password"`
	Email    string `json:"Email"`
}
