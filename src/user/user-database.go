package user

import (
	"agora/src/log"
	"database/sql"
	"errors"
)

const TABLE_QUERY = `CREATE TABLE IF NOT EXISTS users (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		email TEXT UNIQUE NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		);
		`

func (uh *UserHandler) CreateDBTable() error {
	_, err := uh.db.Exec(TABLE_QUERY)
	if err != nil {
		log.Error.Printf("Error creating posts table: %v", err)
		return err
	}
	return nil
}

// InsertNewUser inserts a new user into the database
func (uh *UserHandler) insertNewUser(u User) (int64, error) {
	// Insert a new user into the database
	result, err := uh.db.Exec("INSERT INTO users (id, name, email) VALUES (?, ?, ?)", u.ID, u.Name, u.Email)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

// QueryOneUser queries a user by ID
func (uh *UserHandler) queryOneUser(id string) (User, error) {
	rows, err := uh.db.Query("SELECT id, name, email FROM users WHERE id = ?", id)
	if err != nil {
		return User{}, err
	}
	defer rows.Close()

	var wantedUser User
	for rows.Next() {
		wantedUser, err = scanToUser(rows)
		if err != nil {
			log.Error.Printf("Error scanning user with id=%s: %v", id, err)
			return NullUser, err
		}

		if wantedUser.ID == "" {
			return NullUser, errors.New("user not found")
		}
		return wantedUser, nil
	}

	return NullUser, errors.New("user not found")
}

func (uh *UserHandler) queryAllUsers() ([]User, error) {
	rows, err := uh.db.Query("SELECT id, name, email FROM users")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		user, err := scanToUser(rows)
		if err != nil {
			log.Error.Printf("Error scanning user: %v", err)
			return nil, err
		}
		users = append(users, user)
	}

	if len(users) == 0 {
		return nil, errors.New("no users found")
	}

	return users, nil
}

func scanToUser(rows *sql.Rows) (User, error) {
	var id string
	var name string
	var email string
	if err := rows.Scan(&id, &name, &email); err != nil {
		return NullUser, err
	}
	return User{
		ID:    id,
		Name:  name,
		Email: email,
	}, nil
}
