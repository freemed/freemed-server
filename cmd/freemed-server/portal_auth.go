package main

import (
	"context"
	"crypto/md5"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/freemed/freemed-server/common"
	"github.com/freemed/freemed-server/config"
	"github.com/freemed/freemed-server/dbgen"
	"github.com/freemed/freemed-server/model"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

var (
	portalAuthMiddleware            *jwt.GinJWTMiddleware
	portalAuthMiddlewareInitialized bool
	portalIdentityKey               = "patient_id"
)

// portalLogin represents the JSON body for patient portal login.
type portalLogin struct {
	PatientID   string `json:"patient_id" binding:"required"`
	DateOfBirth string `json:"date_of_birth" binding:"required"`
	Pin         string `json:"pin" binding:"required"`
}

// patientRow holds the subset of patient columns needed for portal auth.
type patientRow struct {
	ID                   int64
	Ptid                 string
	Ptarchive            int64
	Ptdob                sql.NullTime
	PortalPassword       string
	PortalPin            string
	PortalEnabled        bool
	PortalFailedAttempts int
}

// portalPayload is returned from the Authenticator on success.
type portalPayload struct {
	PatientID int64
}

// md5hashPortal produces an MD5 hex string for legacy PIN comparison.
func md5hashPortal(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

// verifyPortalCredentials checks a patient's PIN or password against stored hashes.
// Tries bcrypt against portal_pin, then portal_password, then MD5 fallback.
func verifyPortalCredentials(row patientRow, input string) bool {
	// Try bcrypt against portal_pin first
	if row.PortalPin != "" {
		if err := bcrypt.CompareHashAndPassword([]byte(row.PortalPin), []byte(input)); err == nil {
			return true
		}
	}

	// Try bcrypt against portal_password
	if row.PortalPassword != "" {
		if err := bcrypt.CompareHashAndPassword([]byte(row.PortalPassword), []byte(input)); err == nil {
			return true
		}
	}

	// MD5 fallback for legacy hashes in portal_pin or portal_password
	md5Input := md5hashPortal(input)
	if row.PortalPin == md5Input || row.PortalPassword == md5Input {
		return true
	}

	return false
}

// getPortalPatientByPtid queries the patient table by ptid, excluding archived records.
func getPortalPatientByPtid(ptid string) (patientRow, error) {
	var row patientRow
	err := model.SqlDb.QueryRowContext(context.Background(),
		`SELECT id, ptid, ptarchive, ptdob, portal_password, portal_pin, portal_enabled, portal_failed_attempts
		 FROM patient WHERE ptid = ? AND ptarchive = 0`, ptid).Scan(
		&row.ID, &row.Ptid, &row.Ptarchive, &row.Ptdob,
		&row.PortalPassword, &row.PortalPin, &row.PortalEnabled, &row.PortalFailedAttempts,
	)
	if err != nil {
		return patientRow{}, err
	}
	return row, nil
}

func getPortalAuthMiddleware() *jwt.GinJWTMiddleware {
	if portalAuthMiddlewareInitialized {
		return portalAuthMiddleware
	}

	var err error
	portalAuthMiddleware, err = jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "FreeMED Patient Portal",
		Key:         []byte(config.Config.Session.Key),
		Timeout:     time.Minute * time.Duration(config.Config.Session.Expiry),
		MaxRefresh:  time.Hour,
		IdentityKey: portalIdentityKey,
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginVals portalLogin
			if err := c.ShouldBind(&loginVals); err != nil {
				return nil, jwt.ErrMissingLoginValues
			}

			// Look up patient by ptid
			row, err := getPortalPatientByPtid(loginVals.PatientID)
			if err != nil {
				log.Printf("portalAuthenticator: patient lookup failed for ptid=%s: %v", loginVals.PatientID, err)
				return nil, jwt.ErrFailedAuthentication
			}

			// Check portal_enabled (account lockout)
			if !row.PortalEnabled {
				log.Printf("portalAuthenticator: portal disabled for patient %d (ptid=%s)", row.ID, loginVals.PatientID)
				// Record failed audit log
				_ = insertPortalAuditLog(row.ID, "login", c, false)
				return nil, jwt.ErrFailedAuthentication
			}

			// Verify date of birth
			parsedDOB, err := common.ParseDate(loginVals.DateOfBirth)
			if err != nil {
				log.Printf("portalAuthenticator: invalid DOB format for ptid=%s: %v", loginVals.PatientID, err)
				recordPortalLoginFailure(row, c, loginVals.PatientID)
				return nil, jwt.ErrFailedAuthentication
			}
			if !row.Ptdob.Valid {
				log.Printf("portalAuthenticator: patient %d has no DOB set", row.ID)
				recordPortalLoginFailure(row, c, loginVals.PatientID)
				return nil, jwt.ErrFailedAuthentication
			}
			// Compare dates (ignore time portion)
			ptdobDate := row.Ptdob.Time.Truncate(24 * time.Hour)
			dobDate := parsedDOB.Truncate(24 * time.Hour)
			if !ptdobDate.Equal(dobDate) {
				log.Printf("portalAuthenticator: DOB mismatch for patient %d", row.ID)
				recordPortalLoginFailure(row, c, loginVals.PatientID)
				return nil, jwt.ErrFailedAuthentication
			}

			// Verify PIN/password
			if !verifyPortalCredentials(row, loginVals.Pin) {
				log.Printf("portalAuthenticator: invalid credentials for patient %d", row.ID)
				recordPortalLoginFailure(row, c, loginVals.PatientID)
				return nil, jwt.ErrFailedAuthentication
			}

			// Success: reset failed attempts, update last login, insert audit log
			resetPortalLoginSuccess(row.ID)

			// Insert success audit log
			_ = insertPortalAuditLog(row.ID, "login", c, true)

			log.Printf("portalAuthenticator: successful login for patient %d (ptid=%s)", row.ID, loginVals.PatientID)
			return &portalPayload{PatientID: row.ID}, nil
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			return portalAuthorizeRequest(c, data)
		},
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*portalPayload); ok {
				return jwt.MapClaims{
					portalIdentityKey: v.PatientID,
					"jti":             uuid.New().String(),
					"role":            "patient",
					"portal":          true,
				}
			}
			return jwt.MapClaims{}
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		TokenLookup:   "header:Authorization,cookie:portal_jwt",
		TokenHeadName: "Bearer",
		TimeFunc:      time.Now,
		LoginResponse: func(c *gin.Context, code int, token string, expire time.Time) {
			maxAge := int(time.Until(expire).Seconds())
			if maxAge < 0 {
				maxAge = 0
			}
			secure := config.Config.Web.Keys.Cert != "" && config.Config.Web.Keys.Key != ""
			c.SetCookie("portal_jwt", token, maxAge, "/", "", secure, true)
			c.JSON(http.StatusOK, gin.H{
				"code":    http.StatusOK,
				"expire":  expire.Format(time.RFC3339),
				"message": "login successful",
			})
		},
		RefreshResponse: func(c *gin.Context, code int, token string, t time.Time) {
			maxAge := int(time.Until(t).Seconds())
			if maxAge < 0 {
				maxAge = 0
			}
			secure := config.Config.Web.Keys.Cert != "" && config.Config.Web.Keys.Key != ""
			c.SetCookie("portal_jwt", token, maxAge, "/", "", secure, true)
			c.JSON(http.StatusOK, gin.H{
				"code":    http.StatusOK,
				"token":   token,
				"expire":  t.Format(time.RFC3339),
				"message": "refresh successfully",
			})
		},
	})
	if err != nil {
		panic(err)
	}
	portalAuthMiddlewareInitialized = true
	return portalAuthMiddleware
}

// recordPortalLoginFailure increments portal_failed_attempts and locks the account
// after 5 consecutive failures.
func recordPortalLoginFailure(row patientRow, c *gin.Context, ptid string) {
	newCount := row.PortalFailedAttempts + 1
	enabled := row.PortalEnabled
	if newCount >= 5 {
		enabled = false
	}

	_, err := model.SqlDb.ExecContext(context.Background(),
		`UPDATE patient SET portal_failed_attempts = ?, portal_enabled = ? WHERE id = ?`,
		newCount, enabled, row.ID,
	)
	if err != nil {
		log.Printf("recordPortalLoginFailure: failed to update patient %d: %v", row.ID, err)
	}

	// Insert failure audit log
	_ = insertPortalAuditLog(row.ID, "login", c, false)

	if !enabled {
		log.Printf("portalAuthenticator: account locked for patient %d (ptid=%s) after %d failures", row.ID, ptid, newCount)
	}
}

// resetPortalLoginSuccess resets failed attempts and updates last login time.
func resetPortalLoginSuccess(patientID int64) {
	_, err := model.SqlDb.ExecContext(context.Background(),
		`UPDATE patient SET portal_failed_attempts = 0, portal_last_login = NOW() WHERE id = ?`,
		patientID,
	)
	if err != nil {
		log.Printf("resetPortalLoginSuccess: failed to update patient %d: %v", patientID, err)
	}
}

// insertPortalAuditLog inserts a portal audit log entry using the sqlc-generated query.
func insertPortalAuditLog(patientID int64, action string, c *gin.Context, success bool) error {
	_, err := model.Queries.InsertPortalAuditLog(context.Background(), dbgen.InsertPortalAuditLogParams{
		PatientID: patientID,
		Action:    action,
		IpAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
		Success:   success,
	})
	if err != nil {
		log.Printf("insertPortalAuditLog: failed for patient %d: %v", patientID, err)
	}
	return err
}

// portalAuthorizeRequest validates the JWT token for portal auth.
func portalAuthorizeRequest(c *gin.Context, data interface{}) bool {
	claims := jwt.ExtractClaims(c)
	patientID, ok := claims[portalIdentityKey]
	if !ok {
		return false
	}
	_ = patientID

	// Check if the role is "patient" and portal is true
	role, _ := claims["role"]
	portal, _ := claims["portal"]
	if role != "patient" || portal != true {
		return false
	}

	// Check if this token has been blacklisted
	if jti, ok := claims["jti"]; ok {
		jtiStr := fmt.Sprintf("%v", jti)
		blacklisted, err := common.ActiveSession.IsTokenBlacklisted(jtiStr)
		if err == nil && blacklisted {
			return false
		}
	}

	return true
}

// portalAuthMiddlewareLogout handles portal logout with token blacklisting.
func portalAuthMiddlewareLogout(c *gin.Context) {
	getPortalAuthMiddleware().MiddlewareFunc()(c)

	claims := jwt.ExtractClaims(c)
	if jti, ok := claims["jti"]; ok {
		jtiStr := fmt.Sprintf("%v", jti)
		err := common.ActiveSession.BlacklistToken(jtiStr, time.Minute*time.Duration(config.Config.Session.Expiry))
		if err != nil {
			log.Printf("PortalAuthLogout(): failed to blacklist token %s: %v", jtiStr, err)
		}
	}

	// Clear the httpOnly portal_jwt cookie
	c.SetCookie("portal_jwt", "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "logout successful"})
}

// portalAuthMe returns the current patient's info from JWT claims.
func portalAuthMe(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	patientID, _ := claims[portalIdentityKey]

	// Cast patientID to int64
	var pid int64
	switch v := patientID.(type) {
	case float64:
		pid = int64(v)
	case int64:
		pid = v
	default:
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "message": "invalid patient_id claim"})
		return
	}

	// Look up patient info from DB for the response
	var ptdob sql.NullTime
	var ptfname, ptlname string
	err := model.SqlDb.QueryRowContext(context.Background(),
		`SELECT ptfname, ptlname, ptdob FROM patient WHERE id = ? AND ptarchive = 0`,
		pid,
	).Scan(&ptfname, &ptlname, &ptdob)
	if err != nil {
		log.Printf("portalAuthMe: patient lookup failed for %d: %v", pid, err)
		c.JSON(http.StatusUnauthorized, gin.H{"code": http.StatusUnauthorized, "message": "patient not found"})
		return
	}

	resp := gin.H{
		"patient_id": pid,
		"role":       claims["role"],
		"portal":     claims["portal"],
		"first_name": ptfname,
		"last_name":  ptlname,
	}
	if ptdob.Valid {
		resp["date_of_birth"] = ptdob.Time.Format("2006-01-02")
	}

	c.JSON(http.StatusOK, resp)
}
