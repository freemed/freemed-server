package model

import (
	"database/sql"
	"log"

	"github.com/freemed/freemed-server/common"
)

const (
	TABLE_USER = "user"
)

// UserModel represents a single entry in the user table
type UserModel struct {
	Id                  int64         `db:"id"`
	Username            string        `db:"username"`
	Password            string        `db:"userpassword"`
	Type                NullString    `db:"usertype"`
	ProviderId          int64         `db:"userrealphy"`
	FirstName           NullString    `db:"userfname"`
	MiddleName          NullString    `db:"usermname"`
	LastName            NullString    `db:"userlname"`
	Description         NullString    `db:"userdescrip"`
	Level               []byte        `db:"userlevel"`
	FacilityAccess      []byte        `db:"userfac"`
	ProviderAccess      []byte        `db:"userphy"`
	ProviderGroupAccess []byte        `db:"userphygrp"`
	Options             []byte        `db:"usermanageopt"`
	Email               NullString    `db:"useremail"`
	Sms                 sql.NullInt64 `db:"usersms"`
	SmsProvider         sql.NullInt64 `db:"usersmsprovider"`
	Title               NullString    `db:"usertitle"`
	authenticated       bool          `db:"-"`
}

func init() {
	DbTables = append(DbTables, DbTable{TableName: TABLE_USER, Obj: UserModel{}, Key: "Id"})
}

// Login will preform any actions that are required to make a user model
// officially authenticated.
func (u *UserModel) Login() {
	// Update last login time
	// Add to logged-in user's list
	// etc ...
	u.authenticated = true
}

// Logout will preform any actions that are required to completely
// logout a user.
func (u *UserModel) Logout() {
	// Remove from logged-in user's list
	// etc ...
	u.authenticated = false
}

// IsAuthenticated returns a boolean representing whether the user
// is authenticated.
func (u *UserModel) IsAuthenticated() bool {
	return u.authenticated
}

// UniqueId returns the current object's primary key.
func (u *UserModel) UniqueId() interface{} {
	return u.Id
}

// GetUserByName will populate a user object from a database model with
// a matching user name.
func GetUserByName(username string) (UserModel, error) {
	var u UserModel
	err := DbMap.SelectOne(&u, "SELECT * FROM "+TABLE_USER+" WHERE username = ?", username)
	return u, err
}

// GetUserById will populate a user object from a database model with
// a matching id.
func GetUserById(userId string) (UserModel, error) {
	var u UserModel
	err := DbMap.SelectOne(&u, "SELECT * FROM "+TABLE_USER+" WHERE id = ?", userId)
	return u, err
}

// GetById will populate a user object from a database model with
// a matching id.
func (u *UserModel) GetById(id interface{}) error {
	err := DbMap.SelectOne(u, "SELECT * FROM "+TABLE_USER+" WHERE id = ?", id)
	if err != nil {
		return err
	}

	return nil
}

// CheckUserPassword attempts to authenticate the provided user name and
// password and returns the user id and a boolean representing success.
func CheckUserPassword(username, userpassword string) (int64, bool) {
	u := UserModel{}
	err := DbMap.SelectOne(&u, "SELECT * FROM "+TABLE_USER+" WHERE username = :user AND userpassword = :pass", map[string]interface{}{
		"user": username,
		"pass": common.Md5hash(userpassword),
	})
	if err != nil {
		log.Print(err.Error())
		return 0, false
	}
	if u.Id > 0 {
		return u.Id, true
	}
	return 0, false
}
