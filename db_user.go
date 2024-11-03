package main

import (
	"database/sql"
)


func createUser(db *sql.DB, user User) (int64, error) {
	result, err := db.Exec(queryCreateBuilder(user,"users"), user.UserName, user.Name, user.Email, user.Hash, user.Role)
	if err != nil {
		return 0, err
	} 
	return result.LastInsertId()
}

func getUserById(db *sql.DB, id int) (User, error) {
	var user User
	err := db.QueryRow("SELECT id, username, name, email, hash, role FROM users WHERE id = ?", id).Scan(&user.ID, &user.UserName, &user.Name, &user.Email, &user.Hash, &user.Role)
	return user, err
}

func updateUserById(db *sql.DB, user User) error {
	_, err := db.Exec("UPDATE users SET username = ?, name = ?, email = ?, hash = ?, role = ? WHERE id = ?",
		user.UserName, user.Name, user.Email, user.Hash, user.Role, user.ID)
	return err
}

func deleteUserById(db *sql.DB, id int) error {
	_, err := db.Exec("DELETE FROM users WHERE id = ?", id)
	return err
}
