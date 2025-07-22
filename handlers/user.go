package handlers

import (
	"context"
	"crud-api/models"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
)

// CreateUser handles POST /users
func CreateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var u models.User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		// Check duplicate email
		var exists bool
		if err := db.QueryRowContext(ctx,
			"SELECT EXISTS(SELECT 1 FROM users WHERE email=?)", u.Email).
			Scan(&exists); err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		if exists {
			http.Error(w, "email already exists", http.StatusBadRequest)
			return
		}

		// Insert
		res, err := db.ExecContext(ctx,
			"INSERT INTO users (first_name, last_name, email, password, age) VALUES (?,?,?,?,?)",
			u.FirstName, u.LastName, u.Email, u.Password, u.Age,
		)
		if err != nil {
			http.Error(w, "insert failed", http.StatusInternalServerError)
			return
		}
		id, _ := res.LastInsertId()
		u.ID = int(id)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(u)
	}
}

// GetAllUsers handles GET /users
func GetAllUsers(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT id, first_name, last_name, email, password, age FROM users")
		if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var list []models.User
		for rows.Next() {
			var u models.User
			rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.Password, &u.Age)
			list = append(list, u)
		}
		json.NewEncoder(w).Encode(list)
	}
}

// GetUser handles GET /users/{id}
func GetUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(mux.Vars(r)["id"])
		var u models.User

		err := db.QueryRow(
			"SELECT id, first_name, last_name, email, password, age FROM users WHERE id=?",
			id,
		).Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.Password, &u.Age)

		if err == sql.ErrNoRows {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "db error", http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(u)
	}
}

// UpdateUser handles PUT /users/{id}
func UpdateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(mux.Vars(r)["id"])
		var u models.User
		json.NewDecoder(r.Body).Decode(&u)
		u.ID = id

		// Duplicate email check
		var exists bool
		db.QueryRow(
			"SELECT EXISTS(SELECT 1 FROM users WHERE email=? AND id!=?)",
			u.Email, id,
		).Scan(&exists)
		if exists {
			http.Error(w, "email already in use", http.StatusBadRequest)
			return
		}

		res, err := db.Exec(
			"UPDATE users SET first_name=?, last_name=?, email=?, password=?, age=? WHERE id=?",
			u.FirstName, u.LastName, u.Email, u.Password, u.Age, u.ID,
		)
		if err != nil {
			http.Error(w, "update failed", http.StatusInternalServerError)
			return
		}
		if n, _ := res.RowsAffected(); n == 0 {
			http.Error(w, "no record updated", http.StatusBadRequest)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"message": "updated"})
	}
}

// PatchUser handles PATCH /users/{id}

func PatchUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, _ := strconv.Atoi(mux.Vars(r)["id"])

		// 1. Decode incoming body into map
		var upd map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&upd); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if len(upd) == 0 {
			http.Error(w, "nothing to update", http.StatusBadRequest)
			return
		}

		// 2. Check for duplicate email
		if emailVal, ok := upd["email"].(string); ok {
			var exists bool
			db.QueryRow(
				"SELECT EXISTS(SELECT 1 FROM users WHERE email=? AND id!=?)",
				emailVal, id,
			).Scan(&exists)
			if exists {
				http.Error(w, "email already exists", http.StatusBadRequest)
				return
			}
		}

		// 3. Build dynamic SQL SET clause
		var setParts []string
		var args []interface{}
		for key, val := range upd {
			setParts = append(setParts, fmt.Sprintf("%s = ?", key))
			args = append(args, val)
		}
		args = append(args, id)

		// 4. Final query & execution
		query := fmt.Sprintf("UPDATE users SET %s WHERE id = ?", strings.Join(setParts, ", "))
		res, err := db.Exec(query, args...)
		if err != nil {
			http.Error(w, "patch failed", http.StatusInternalServerError)
			return
		}
		if count, _ := res.RowsAffected(); count == 0 {
			http.Error(w, "no record patched", http.StatusBadRequest)
			return
		}

		json.NewEncoder(w).Encode(map[string]string{"message": "patched"})
	}
}

// DeleteUser handles DELETE /users/{id}

func DeleteUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(mux.Vars(r)["id"])
		if err != nil {
			http.Error(w, "invalid user id", http.StatusBadRequest)
			return
		}

		res, err := db.Exec("DELETE FROM users WHERE id = ?", id)
		if err != nil {
			http.Error(w, "delete failed", http.StatusInternalServerError)
			return
		}

		n, err := res.RowsAffected()
		if err != nil {
			http.Error(w, "cannot get affected rows", http.StatusInternalServerError)
			return
		}

		if n == 0 {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}

		// Send JSON response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "User deleted successfully",
		})
	}
}
