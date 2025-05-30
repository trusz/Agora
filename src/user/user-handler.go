package user

import (
	"agora/src/db"
	"agora/src/log"
)

type UserHandler struct {
	db *db.DB
}

func NewUserHandler(db *db.DB) *UserHandler {
	return &UserHandler{
		db: db,
	}
}

func (uh *UserHandler) AddUser(id, name, email string) (User, error) {
	user := User{
		ID:    id,
		Name:  name,
		Email: email,
	}
	if _, err := uh.insertNewUser(user); err != nil {
		log.Error.Printf("Error adding user: %v", err)
		return NullUser, err
	}
	return user, nil
}

func (uh *UserHandler) UserExists(id string) bool {
	user, err := uh.queryOneUser(id)
	if err != nil {
		return false
	}
	return user.ID != ""
}

func (uh *UserHandler) RetrieveUserMap() (map[string]User, error) {
	users, err := uh.queryAllUsers()
	if err != nil {
		log.Error.Printf("Error retrieving users: %v", err)
		return nil, err
	}
	return uh.userSliceToMap(users), nil
}

func (uh *UserHandler) userSliceToMap(users []User) map[string]User {
	userMap := make(map[string]User)
	for _, user := range users {
		userMap[user.ID] = user
	}
	return userMap
}
