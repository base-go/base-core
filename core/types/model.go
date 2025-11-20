package types

type UserData struct {
	Id        uint   `json:"id"`
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	Email     string `json:"email"`
	Username  string `json:"username"`

	// Add other necessary fields
}
