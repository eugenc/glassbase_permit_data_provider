package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
	Name      string `json:"name"`
	Role      string `json:"role"`
}

func HandleLogin(pool *pgxpool.Pool, jwtSecret []byte) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req loginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}

		var id int
		var hash, name, role string
		err := pool.QueryRow(r.Context(),
			`SELECT id, password_hash, name, role FROM admin_users WHERE email = $1`,
			req.Email).Scan(&id, &hash, &name, &role)
		if err != nil {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		_, _ = pool.Exec(r.Context(),
			`UPDATE admin_users SET last_login_at = NOW() WHERE id = $1`, id)

		exp := time.Now().Add(24 * time.Hour)
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"sub":  strconv.Itoa(id),
			"name": name,
			"role": role,
			"exp":  exp.Unix(),
		})
		signed, err := token.SignedString(jwtSecret)
		if err != nil {
			http.Error(w, "token error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(loginResponse{
			Token:     signed,
			ExpiresAt: exp.Unix(),
			Name:      name,
			Role:      role,
		})
	}
}
