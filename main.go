// This is a Go file
package main

import (
	"database/sql"
	f "fmt"
	"html/template"
	"net/http"
	"unicode"

	"golang.org/x/crypto/bcrypt"

	_ "github.com/go-sql-driver/mysql"
)

var tpl *template.Template
var db *sql.DB

func main() {
	tpl, _ = template.ParseGlob("templates/*.html")
	var err error
	db, err = sql.Open("mysql", "root:Exploring here 55!@tcp(localhost:3306)/public")

	if err != nil {
		panic(err.Error())
	}
	defer db.Close()

	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/registerauth", registerAuthHandler)
	http.ListenAndServe("localhost:8080", nil)
}

// Serve the form for registering new users
func registerHandler(w http.ResponseWriter, r *http.Request) {
	f.Println("registerHandler is running")
	tpl.ExecuteTemplate(w, "index.html", nil)
}

// Create new user in database
func registerAuthHandler(w http.ResponseWriter, r *http.Request) {
	f.Println("registerAuthHandler is running")
	// Parse the form
	r.ParseForm()
	username := r.FormValue("username")

	// Check username for only alphanumeric characters
	var alphanumericName = true

	// Go through the username characters
	for _, char := range username {
		// Check if the character is a letter or a number
		if !unicode.IsLetter(char) && !unicode.IsNumber(char) {
			alphanumericName = false
		}
	}

	// Check if the name is between 5 & 50 characters
	var nameLength bool

	if 5 <= len(username) && len(username) <= 50 {
		nameLength = true
	}

	// Check the password
	password := r.FormValue("password")
	var passwordLowerCase, passwordUpperCase, passwordNumber, passwordSpecial, passwordLength, passwordNoSpaces bool
	passwordNoSpaces = true

	for _, char := range password {
		switch {
		case unicode.IsLower(char):
			passwordLowerCase = true
		case unicode.IsUpper(char):
			passwordUpperCase = true
		case unicode.IsNumber(char):
			passwordNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			passwordSpecial = true
		case unicode.IsSpace(int32(char)):
			passwordNoSpaces = false
		}
	}

	if 5 < len(password) && len(password) < 50 {
		passwordLength = true
	}

	f.Println("alphanumericName: ", alphanumericName, "\nnameLength: ", nameLength, "\npasswordLength: ", passwordLength, "\npasswordLowerCase: ", passwordLowerCase, "\npasswordUpperCase: ", passwordUpperCase, "\npasswordNumber: ", passwordNumber, "\npasswordSpecial: ", passwordSpecial, "\npasswordLength: ", passwordLength, "\npasswordNoSpaces: ", passwordNoSpaces)
	if !alphanumericName || !nameLength || !passwordLowerCase || !passwordUpperCase || !passwordNumber || !passwordSpecial || !passwordLength || !passwordNoSpaces {
		tpl.ExecuteTemplate(w, "index.html", "Invalid password")
		return
	}

	// Check if the username already exists
	statement := "SELECT id FROM users WHERE username = ?"
	row := db.QueryRow(statement, username)

	var userID string

	err := row.Scan(&userID)

	if err != sql.ErrNoRows {
		f.Println("The username already exists.")
		f.Println("Error: ", err)

		tpl.ExecuteTemplate(w, "index.html", "The username has already been taken")
		return
	}

	// Create a hash form of the password
	var hash []byte

	hash, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		f.Println("bcrypt error: ", err)
		tpl.ExecuteTemplate(w, "index.html", "An issue was encountered during the account registration")
		return
	}

	var insertStatement *sql.Stmt
	insertStatement, err = db.Prepare("INSERT INTO users (username, hash) VALUES (?, ?);")

	if err != nil {
		f.Println("Error preparing the statement: ", err)
		tpl.ExecuteTemplate(w, "index.html", "An issue was encountered during the account registration")
		return
	}
	defer insertStatement.Close()

	var result sql.Result

	result, err = insertStatement.Exec(username, hash)

	f.Println("Result: ", result)

	if err != nil {
		f.Println("Error inserting a new user")
		tpl.ExecuteTemplate(w, "index.html", "An issue was encountered during the account registration")
		return
	}
	f.Fprint(w, "Account successfully created")
}