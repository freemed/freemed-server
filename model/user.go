package model

import (
	"database/sql"
	"time"
	"context"
	"log"
	"strconv"

	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/dbgen"
)

const (
	TABLE_USER = "user"
)

// UserModel represents a single entry in the user table
type UserModel struct {
	ID        int64          `db:"id" json:"id"`
	CreatedAt time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt sql.NullTime   `db:"deleted_at" json:"deleted_at"`
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
	Sms                 NullInt64     `db:"usersms"`
	SmsProvider         NullInt64     `db:"usersmsprovider"`
	Title               NullString    `db:"usertitle"`
	authenticated       bool          `db:"-"`
}

// TableName overrides the table name used by User to `profiles`
func (UserModel) TableName() string {
	return TABLE_USER
}

func init() {
}

// Login will preform any actions that are required to make a user model
// officially authenticated.
func (u *UserModel) Login() {
	u.authenticated = true
}

// Logout will preform any actions that are required to completely
// logout a user.
func (u *UserModel) Logout() {
	u.authenticated = false
}

// IsAuthenticated returns a boolean representing whether the user
// is authenticated.
func (u *UserModel) IsAuthenticated() bool {
	return u.authenticated
}

// UniqueId returns the current object's primary key.
func (u *UserModel) UniqueId() interface{} {
	return u.ID
}

// GetUserByName will populate a user object from a database model with
// a matching user name.
func GetUserByName(username string) (dbgen.User, error) {
	return Queries.GetUserByUsername(context.Background(), username)
}

// GetUserById will populate a user object from a database model with
// a matching id.
func GetUserById(userId string) (dbgen.User, error) {
	id, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		return dbgen.User{}, err
	}
	return Queries.GetUserById(context.Background(), id)
}

// GetById will populate a user object from a database model with
// a matching id.
func (u *UserModel) GetById(id interface{}) error {
	return nil // deprecated: use GetUserById
}

// CheckUserPassword attempts to authenticate the provided user name and
// password and returns the user id and a boolean representing success.
func CheckUserPassword(username, userpassword string) (int64, bool) {
	u, err := Queries.CheckUserPassword(context.Background(), dbgen.CheckUserPasswordParams{
		Username: username,
		Password: common.Md5hash(userpassword),
	})
	if err != nil {
		log.Print(err.Error())
		return 0, false
	}

	if u.ID > 0 {
		return u.ID, true
	}
	return 0, false
}
