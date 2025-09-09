package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var db *sql.DB

const (
	sessionCookieName = "session_id"
	sessionDuration   = 24 * time.Hour
)

type User struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", "./streamapp.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create tables if not exist
	createTables()

	http.HandleFunc("/api/register", cors(registerHandler))
	http.HandleFunc("/api/login", cors(loginHandler))
	http.HandleFunc("/api/logout", cors(authRequired(logoutHandler)))
	http.HandleFunc("/api/me", cors(authRequired(meHandler)))
	http.HandleFunc("/api/start-stream", cors(requireRole("player", startStreamHandler)))

	fmt.Println("Server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// ---------------- tables ----------------
func createTables() {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		role TEXT NOT NULL
	);`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS sessions (
		id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		expires_at DATETIME NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(id)
	);`)
	if err != nil {
		log.Fatal(err)
	}
}

// ---------------- middleware ----------------
func cors(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		h(w, r)
	}
}

// ---------------- register/login ----------------
type registerReq struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req registerReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	allowed := map[string]bool{"player": true, "watcher": true}
	if !allowed[req.Role] {
		http.Error(w, "invalid role", http.StatusBadRequest)
		return
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	res, err := db.Exec(
		"INSERT INTO users (username,email,password_hash,role) VALUES (?,?,?,?)",
		req.Username, req.Email, string(hash), req.Role,
	)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "could not create user", http.StatusBadRequest)
		return
	}

	id, _ := res.LastInsertId()
	user := User{ID: int(id), Username: req.Username, Email: req.Email, Role: req.Role}
	json.NewEncoder(w).Encode(user)
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req loginReq
	json.NewDecoder(r.Body).Decode(&req)

	var id int
	var hash, username, role string
	err := db.QueryRow("SELECT id, password_hash, username, role FROM users WHERE email=?", req.Email).
		Scan(&id, &hash, &username, &role)
	if err != nil || bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)) != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	sid := uuid.New().String()
	exp := time.Now().Add(sessionDuration)
	db.Exec("INSERT INTO sessions (id, user_id, expires_at) VALUES (?,?,?)", sid, id, exp)

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sid,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  exp,
	})

	user := User{ID: id, Username: username, Email: req.Email, Role: role}
	json.NewEncoder(w).Encode(user)
}

// ---------------- auth/session ----------------
type authHandler func(w http.ResponseWriter, r *http.Request, user User)

func authRequired(next authHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := userFromRequest(r)
		if err != nil {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}
		next(w, r, user)
	}
}

func userFromRequest(r *http.Request) (User, error) {
	c, err := r.Cookie(sessionCookieName)
	if err != nil {
		return User{}, errors.New("no cookie")
	}
	var uid int
	var expires time.Time
	err = db.QueryRow("SELECT user_id, expires_at FROM sessions WHERE id=?", c.Value).Scan(&uid, &expires)
	if err != nil || time.Now().After(expires) {
		return User{}, errors.New("invalid session")
	}
	var u User
	db.QueryRow("SELECT id, username, email, role FROM users WHERE id=?", uid).
		Scan(&u.ID, &u.Username, &u.Email, &u.Role)
	return u, nil
}

func logoutHandler(w http.ResponseWriter, r *http.Request, user User) {
	c, err := r.Cookie(sessionCookieName)
	if err == nil {
		db.Exec("DELETE FROM sessions WHERE id=?", c.Value)
		http.SetCookie(w, &http.Cookie{
			Name:     sessionCookieName,
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Unix(0, 0),
		})
	}
	w.Write([]byte("logged out"))
}

func meHandler(w http.ResponseWriter, r *http.Request, user User) {
	json.NewEncoder(w).Encode(user)
}

// ---------------- role middleware ----------------
func requireRole(role string, next authHandler) http.HandlerFunc {
	return authRequired(func(w http.ResponseWriter, r *http.Request, user User) {
		if user.Role != role {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next(w, r, user)
	})
}

// ---------------- player endpoint ----------------
func startStreamHandler(w http.ResponseWriter, r *http.Request, user User) {
	w.Write([]byte("Stream started by " + user.Username))
}
