package model

import (
	"context"
	"crypto/md5"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/freemed/freemed-server/dbgen"
	"golang.org/x/crypto/bcrypt"
)

const (
	TABLE_USER = "user"
)

// UserModel represents a single entry in the user table
type UserModel struct {
	ID                 int64          `db:"id" json:"id"`
	CreatedAt          time.Time      `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time      `db:"updated_at" json:"updated_at"`
	DeletedAt          sql.NullTime   `db:"deleted_at" json:"deleted_at"`
	Username           string         `db:"username"`
	Password           string         `db:"userpassword"`
	Type               NullString     `db:"usertype"`
	ProviderId         int64          `db:"userrealphy"`
	FirstName          NullString     `db:"userfname"`
	MiddleName         NullString     `db:"usermname"`
	LastName           NullString     `db:"userlname"`
	Description        NullString     `db:"userdescrip"`
	Level              []byte         `db:"userlevel"`
	FacilityAccess     []byte         `db:"userfac"`
	ProviderAccess     []byte         `db:"userphy"`
	ProviderGroupAccess []byte        `db:"userphygrp"`
	Options            []byte         `db:"usermanageopt"`
	Email              NullString     `db:"useremail"`
	Sms                NullInt64      `db:"usersms"`
	SmsProvider        NullInt64      `db:"usersmsprovider"`
	Title              NullString     `db:"usertitle"`
	authenticated      bool           `db:"-"`
}

// TableName overrides the table name used by User
func (UserModel) TableName() string {
	return TABLE_USER
}

// Login marks the user as authenticated.
func (u *UserModel) Login() {
	u.authenticated = true
}

// Logout marks the user as not authenticated.
func (u *UserModel) Logout() {
	u.authenticated = false
}

// IsAuthenticated returns whether the user is authenticated.
func (u *UserModel) IsAuthenticated() bool {
	return u.authenticated
}

// UniqueId returns the current object's primary key.
func (u *UserModel) UniqueId() interface{} {
	return u.ID
}

// GetUserByName returns a user by username.
func GetUserByName(username string) (dbgen.User, error) {
	return Queries.GetUserByUsername(context.Background(), username)
}

// GetUserById returns a user by ID string.
func GetUserById(userId string) (dbgen.User, error) {
	id, err := strconv.ParseInt(userId, 10, 64)
	if err != nil {
		return dbgen.User{}, err
	}
	return Queries.GetUserById(context.Background(), id)
}

// GetById is deprecated; use GetUserById.
func (u *UserModel) GetById(id interface{}) error {
	return nil
}

// md5hash produces an MD5 hex string (legacy password comparison only).
func md5hash(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

// CheckUserPassword authenticates a user by username and plaintext password.
// Supports both bcrypt (current) and MD5 (legacy) hashes.
// On legacy MD5 match, the password is upgraded to bcrypt automatically.
func CheckUserPassword(username, password string) (int64, bool) {
	u, err := Queries.GetUserByUsername(context.Background(), username)
	if err != nil {
		log.Printf("CheckUserPassword: user not found: %s", err.Error())
		return 0, false
	}

	// Prefer userpassword_bcrypt when set (bcrypt hash)
	if u.UserpasswordBcrypt != "" {
		if err := bcrypt.CompareHashAndPassword([]byte(u.UserpasswordBcrypt), []byte(password)); err == nil {
			return u.ID, true
		}
		return 0, false
	}

	// Fallback: legacy userpassword column (bcrypt or MD5)
	storedHash := u.Userpassword

	// Try bcrypt first
	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)); err == nil {
		return u.ID, true
	}

	// Fallback: legacy MD5 hash
	if storedHash == md5hash(password) {
		// Upgrade to bcrypt in the background (best-effort)
		go upgradePasswordHash(u.ID, password)
		return u.ID, true
	}

	return 0, false
}

// upgradePasswordHash writes the bcrypt hash to userpassword_bcrypt.
func upgradePasswordHash(userID int64, password string) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("upgradePasswordHash: bcrypt error for user %d: %s", userID, err.Error())
		return
	}
	// Use raw SQL via SqlDb since we don't have an update-password sqlc query
	_, err = SqlDb.ExecContext(context.Background(),
		"UPDATE user SET userpassword_bcrypt = ? WHERE id = ?",
		string(hash), userID)
	if err != nil {
		log.Printf("upgradePasswordHash: update error for user %d: %s", userID, err.Error())
		return
	}
	log.Printf("upgradePasswordHash: upgraded user %d from MD5 to bcrypt", userID)
}

// HashPassword generates a bcrypt hash for a plaintext password.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}
