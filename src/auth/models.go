// User represents an authenticated user
type User struct {
	ID       string
	Name     string
	Email    string
	Username string
	Token    string
}

// UserStore is an interface for storing and retrieving user data
type UserStore interface {
	GetUserByEmail(email string) (*User, error)
	CreateUser(user *User) error
	UpdateUser(user *User) error
}
