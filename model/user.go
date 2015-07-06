package model

import (
	"database/sql"
	"github.com/freemed/freemed-server/common"
	"github.com/martini-contrib/sessionauth"
	"log"
)

const (
	TABLE_USER = "user"
)

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

// GetAnonymousUser should generate an anonymous user model
// for all sessions. This should be an unauthenticated 0 value struct.
func GenerateAnonymousUser() sessionauth.User {
	return &UserModel{}
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

func (u *UserModel) IsAuthenticated() bool {
	return u.authenticated
}

func (u *UserModel) UniqueId() interface{} {
	return u.Id
}

// GetById will populate a user object from a database model with
// a matching id.
func (u *UserModel) GetById(id interface{}) error {
	err := DbMap.SelectOne(u, "SELECT * FROM "+TABLE_USER+" WHERE id = $1", id)
	if err != nil {
		return err
	}

	return nil
}

func CheckUserPassword(username, userpassword string) (int64, bool) {
	u := &UserModel{}
	err := DbMap.SelectOne(u, "SELECT * FROM "+TABLE_USER+" WHERE username = :user AND userpassword = :pass", map[string]interface{}{
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
