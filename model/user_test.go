package model

import (
	"crypto/md5"
	"database/sql/driver"
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/freemed/freemed-server/dbgen"
	"golang.org/x/crypto/bcrypt"
)

// setupMockDB creates a sqlmock DB, wires it into the model package globals,
// and returns a cleanup function that restores the original values.
func setupMockDB(t *testing.T) (sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New(
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp),
	)
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	// Save original globals
	origQueries := Queries
	origSqlDb := SqlDb

	// Wire in mocks
	Queries = dbgen.New(db)
	SqlDb = db

	cleanup := func() {
		Queries = origQueries
		SqlDb = origSqlDb
		db.Close()
	}

	return mock, cleanup
}

// userRow returns a slice of column values matching the GetUserByUsername scan order:
// id, created_at, updated_at, deleted_at, username, userpassword, userpassword_bcrypt,
// usertype, userrealphy, userfname, usermname, userlname, userdescrip, userlevel,
// userfac, userphy, userphygrp, usermanageopt, useremail, usersms, usersmsprovider, usertitle
func userRow(id int64, username, userpassword, userpasswordBcrypt string) []driver.Value {
	return []driver.Value{
		id,                                 // id
		time.Now(),                         // created_at
		time.Now(),                         // updated_at
		nil,                                 // deleted_at (nullable)
		username,                            // username
		userpassword,                        // userpassword
		userpasswordBcrypt,                  // userpassword_bcrypt
		nil,                                 // usertype (nullable)
		int64(0),                            // userrealphy
		nil,                                 // userfname (nullable)
		nil,                                 // usermname (nullable)
		nil,                                 // userlname (nullable)
		nil,                                 // userdescrip (nullable)
		nil,                                 // userlevel (nullable)
		nil,                                 // userfac (nullable)
		nil,                                 // userphy (nullable)
		nil,                                 // userphygrp (nullable)
		nil,                                 // usermanageopt (nullable)
		nil,                                 // useremail (nullable)
		nil,                                 // usersms (nullable)
		nil,                                 // usersmsprovider (nullable)
		nil,                                 // usertitle (nullable)
	}
}

// userColumns returns the column names for GetUserByUsername rows.
func userColumns() []string {
	return []string{
		"id", "created_at", "updated_at", "deleted_at", "username",
		"userpassword", "userpassword_bcrypt", "usertype", "userrealphy",
		"userfname", "usermname", "userlname", "userdescrip", "userlevel",
		"userfac", "userphy", "userphygrp", "usermanageopt", "useremail",
		"usersms", "usersmsprovider", "usertitle",
	}
}

// TestHashPassword verifies that HashPassword generates a valid bcrypt hash.
func TestHashPassword(t *testing.T) {
	password := "mysecret"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword returned empty hash")
	}
	if hash == password {
		t.Fatal("HashPassword returned plaintext, not a hash")
	}

	// Verify the hash is valid bcrypt that matches the original password.
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)); err != nil {
		t.Fatalf("bcrypt.CompareHashAndPassword failed on generated hash: %v", err)
	}
}

// TestCheckUserPassword_Bcrypt verifies authentication when the user has a bcrypt hash
// in userpassword_bcrypt.
func TestCheckUserPassword_Bcrypt(t *testing.T) {
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	password := "correct-password"
	bcryptHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to generate bcrypt hash: %v", err)
	}

	// Test: correct password
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT `)).
		WithArgs("testuser").
		WillReturnRows(sqlmock.NewRows(userColumns()).
			AddRow(userRow(42, "testuser", "", string(bcryptHash))...))

	id, ok := CheckUserPassword("testuser", password)
	if !ok {
		t.Error("CheckUserPassword should succeed with correct password")
	}
	if id != 42 {
		t.Errorf("expected user ID 42, got %d", id)
	}

	// Test: wrong password
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT `)).
		WithArgs("testuser").
		WillReturnRows(sqlmock.NewRows(userColumns()).
			AddRow(userRow(42, "testuser", "", string(bcryptHash))...))

	id, ok = CheckUserPassword("testuser", "wrong-password")
	if ok {
		t.Error("CheckUserPassword should fail with wrong password")
	}
	if id != 0 {
		t.Errorf("expected user ID 0 on failure, got %d", id)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestCheckUserPassword_BcryptInUserpassword verifies authentication when the
// bcrypt hash is in the legacy userpassword column (no userpassword_bcrypt set).
func TestCheckUserPassword_BcryptInUserpassword(t *testing.T) {
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	password := "legacy-pass"
	bcryptHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to generate bcrypt hash: %v", err)
	}

	// userpassword_bcrypt is empty, bcrypt hash is in userpassword column
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT `)).
		WithArgs("legacyuser").
		WillReturnRows(sqlmock.NewRows(userColumns()).
			AddRow(userRow(7, "legacyuser", string(bcryptHash), "")...))

	id, ok := CheckUserPassword("legacyuser", password)
	if !ok {
		t.Error("CheckUserPassword should succeed when bcrypt hash is in userpassword column")
	}
	if id != 7 {
		t.Errorf("expected user ID 7, got %d", id)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestCheckUserPassword_Md5Legacy verifies the MD5 fallback path with a correct password.
func TestCheckUserPassword_Md5Legacy(t *testing.T) {
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	password := "legacy-md5-pass"
	md5Hash := fmt.Sprintf("%x", md5.Sum([]byte(password)))

	// userpassword contains the MD5 hash; userpassword_bcrypt is empty.
	// Set the SELECT expectation first since CheckUserPassword calls GetUserByUsername first.
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT `)).
		WithArgs("md5user").
		WillReturnRows(sqlmock.NewRows(userColumns()).
			AddRow(userRow(99, "md5user", md5Hash, "")...))

	// upgradePasswordHash runs in a goroutine after MD5 match.
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE user SET userpassword_bcrypt`)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	id, ok := CheckUserPassword("md5user", password)
	if !ok {
		t.Error("CheckUserPassword should succeed with MD5 fallback")
	}
	if id != 99 {
		t.Errorf("expected user ID 99, got %d", id)
	}

	// Wait briefly for the async upgradePasswordHash goroutine to run.
	time.Sleep(50 * time.Millisecond)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestCheckUserPassword_Md5LegacyWrongPassword verifies the MD5 fallback
// fails with an incorrect password.
func TestCheckUserPassword_Md5LegacyWrongPassword(t *testing.T) {
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	password := "legacy-md5-pass"
	md5Hash := fmt.Sprintf("%x", md5.Sum([]byte(password)))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT `)).
		WithArgs("md5user").
		WillReturnRows(sqlmock.NewRows(userColumns()).
			AddRow(userRow(99, "md5user", md5Hash, "")...))

	id, ok := CheckUserPassword("md5user", "wrong-password")
	if ok {
		t.Error("CheckUserPassword should fail with wrong MD5 password")
	}
	if id != 0 {
		t.Errorf("expected user ID 0 on failure, got %d", id)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestBcryptUpgrade_TriggersOnMd5Login verifies that a successful MD5 login
// triggers the upgrade goroutine, which calls UPDATE on SqlDb with a bcrypt hash.
func TestBcryptUpgrade_TriggersOnMd5Login(t *testing.T) {
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	password := "upgrade-me"
	md5Hash := fmt.Sprintf("%x", md5.Sum([]byte(password)))

	// SELECT expectation for GetUserByUsername
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT `)).
		WithArgs("md5user").
		WillReturnRows(sqlmock.NewRows(userColumns()).
			AddRow(userRow(99, "md5user", md5Hash, "")...))

	// Expect the async upgrade: UPDATE user SET userpassword_bcrypt = ? WHERE id = ?
	// Use AnyArg for the bcrypt hash (it varies per run) and 99 for the user ID.
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE user SET userpassword_bcrypt = ? WHERE id = ?`)).
		WithArgs(sqlmock.AnyArg(), int64(99)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	id, ok := CheckUserPassword("md5user", password)
	if !ok {
		t.Error("CheckUserPassword should succeed with MD5 fallback")
	}
	if id != 99 {
		t.Errorf("expected user ID 99, got %d", id)
	}

	// Wait for the async upgradePasswordHash goroutine to consume its expectation.
	time.Sleep(50 * time.Millisecond)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestBcryptUpgrade_SkipsForBcrypt verifies that when the user already has a
// bcrypt hash in userpassword_bcrypt, no upgrade UPDATE is issued.
func TestBcryptUpgrade_SkipsForBcrypt(t *testing.T) {
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	password := "already-bcrypt"
	bcryptHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to generate bcrypt hash: %v", err)
	}

	// Only the SELECT should be called — no ExpectExec because no upgrade should happen.
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT `)).
		WithArgs("bcryptuser").
		WillReturnRows(sqlmock.NewRows(userColumns()).
			AddRow(userRow(42, "bcryptuser", "", string(bcryptHash))...))

	id, ok := CheckUserPassword("bcryptuser", password)
	if !ok {
		t.Error("CheckUserPassword should succeed with bcrypt hash")
	}
	if id != 42 {
		t.Errorf("expected user ID 42, got %d", id)
	}

	// Wait briefly to let any stray goroutine execute (none should be spawned).
	time.Sleep(50 * time.Millisecond)

	// ExpectationsWereMet must pass — the only expectation was the SELECT.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations (unexpected upgrade UPDATE?): %v", err)
	}
}

// TestBcryptUpgrade_WritesToNewColumn verifies that the upgrade UPDATE targets
// the userpassword_bcrypt column, not the legacy userpassword column.
func TestBcryptUpgrade_WritesToNewColumn(t *testing.T) {
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	password := "col-test"
	md5Hash := fmt.Sprintf("%x", md5.Sum([]byte(password)))

	// SELECT expectation for GetUserByUsername
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT `)).
		WithArgs("coluser").
		WillReturnRows(sqlmock.NewRows(userColumns()).
			AddRow(userRow(7, "coluser", md5Hash, "")...))

	// The UPDATE must reference userpassword_bcrypt in the SET clause.
	// The regex explicitly captures the full statement to prove the column name.
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE user SET userpassword_bcrypt = ? WHERE id = ?`)).
		WithArgs(sqlmock.AnyArg(), int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	id, ok := CheckUserPassword("coluser", password)
	if !ok {
		t.Error("CheckUserPassword should succeed with MD5 fallback")
	}
	if id != 7 {
		t.Errorf("expected user ID 7, got %d", id)
	}

	// Wait for the async upgradePasswordHash goroutine.
	time.Sleep(50 * time.Millisecond)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}

// TestCheckUserPassword_UserNotFound verifies that a missing user returns (0, false).
func TestCheckUserPassword_UserNotFound(t *testing.T) {
	mock, cleanup := setupMockDB(t)
	defer cleanup()

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT `)).
		WithArgs("nosuchuser").
		WillReturnError(fmt.Errorf("sql: no rows in result set"))

	id, ok := CheckUserPassword("nosuchuser", "any-password")
	if ok {
		t.Error("CheckUserPassword should fail when user not found")
	}
	if id != 0 {
		t.Errorf("expected user ID 0 on not-found, got %d", id)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet expectations: %v", err)
	}
}
