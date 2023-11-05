package app

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hy00nc/conduit-go/internal/database"
	"github.com/hy00nc/conduit-go/internal/utils"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var test_db *gorm.DB

var tests = []struct {
	name               string
	url                string
	init_func          func(*http.Request)
	method             string
	body               string
	expectedStatusCode int
	responseBodyRegex  string
}{
	/* User register tests */
	{
		"Register user",
		"/api/users",
		func(req *http.Request) {},
		"POST",
		`{"user":{"username":"sally","email":"sally@something","password":"strongpassword"}}`,
		http.StatusCreated,
		fmt.Sprintf(`{"user":{"email":"sally@something","token":"([a-zA-Z0-9-_.]+)","username":"sally","bio":"","image":"%s"}}`, defaultImage),
	},
    {
		"Register user",
		"/api/users",
		func(req *http.Request) {},
		"POST",
		`{"user":{"username":"harry","email":"harry@something","password":"strongpassword"}}`,
		http.StatusCreated,
		fmt.Sprintf(`{"user":{"email":"harry@something","token":"([a-zA-Z0-9-_.]+)","username":"harry","bio":"","image":"%s"}}`, defaultImage),
	},
    /* User login tests */
	{
		"User login (normal)",
		"/api/users/login",
		func(req *http.Request) {},
		"POST",
		`{"user":{"username":"sally","email":"sally@something","password":"strongpassword"}}`,
		http.StatusOK,
		fmt.Sprintf(`{"user":{"email":"sally@something","token":"([a-zA-Z0-9-_.]+)","username":"sally","bio":"","image":"%s"}}`, defaultImage),
	},
    {
		"User login (wrong email)",
		"/api/users/login",
		func(req *http.Request) {},
		"POST",
		`{"user":{"email":"jake@something","password":"strongpassword"}}`,
		http.StatusForbidden,
		`{"errors":{"email or password":"is invalid"}}`,
	},
    {
		"User login (wrong password)",
		"/api/users/login",
		func(req *http.Request) {},
		"POST",
		`{"user":{"email":"sally@something","password":"weakpassword"}}`,
		http.StatusForbidden,
		`{"errors":{"email or password":"is invalid"}}`,
	},
	/* Tests with authorization */
	{
		"Get current user",
		"/api/user",
		func(req *http.Request) {
			token, _ := utils.GetToken(1)  // haeyoon
			req.Header.Set("Authorization", fmt.Sprintf("Token %v", token))
		},
		"GET",
		``,
		http.StatusOK,
		``,
	},
    {
		"Get current user with wrong user id",
		"/api/user",
		func(req *http.Request) {
			token, _ := utils.GetToken(99)  // non-existent user
			req.Header.Set("Authorization", fmt.Sprintf("Token %v", token))
		},
		"GET",
		``,
		http.StatusUnauthorized,
		``,
	},
}

// e2e test using table-driven tests (https://github.com/golang/go/wiki/TableDrivenTests)
func TestRoutesAndHandlers(t *testing.T) {
	log.SetOutput(io.Discard)
	// Initialize test DB before running tests
	test_db = database.InitTestDB()
	database.MigrateDB(test_db)
	defer database.RemoveDB(test_db)

	// Setup Router
	r := MakeWebHandler(false)

	// Run tests on test suite
	asserts := assert.New(t)
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(tt.method, tt.url, bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			asserts.NoError(err)

			tt.init_func(req)

			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			asserts.Equal(tt.expectedStatusCode, w.Code, fmt.Sprintf("Should return %d", tt.expectedStatusCode))
			asserts.Regexp(tt.responseBodyRegex, strings.TrimSpace(w.Body.String()), "Response body does not match")
		})
	}
}
