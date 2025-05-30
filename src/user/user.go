package user

type User struct {
	ID    string
	Name  string
	Email string
}

var NullUser = User{
	ID:    "null",
	Name:  "null",
	Email: "null",
}
