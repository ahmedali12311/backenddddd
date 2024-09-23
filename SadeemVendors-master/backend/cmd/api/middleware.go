package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"project/internal/data"
	"project/utils"
	"project/utils/validator"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

type contextKey string

const UserIDKey contextKey = "userID"
const UserRoleKey contextKey = "userRole"

func (app *application) AuthMiddleware(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.jwtErrorResponse(w, r, utils.ErrMissingToken)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			app.jwtErrorResponse(w, r, utils.ErrInvalidToken)
			return
		}

		tokenString := parts[1]
		token, err := utils.ValidateToken(tokenString)
		if err != nil {
			switch err.Error() {
			case "token contains an invalid number of segments":
				app.jwtErrorResponse(w, r, utils.ErrInvalidToken)
			default:
				app.jwtErrorResponse(w, r, utils.ErrInvalidClaims)
			}
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			app.jwtErrorResponse(w, r, utils.ErrInvalidClaims)
			return
		}

		if exp, ok := claims["exp"].(float64); ok {
			expTime := time.Unix(int64(exp), 0)
			if expTime.Before(time.Now()) {
				app.jwtErrorResponse(w, r, utils.ErrExpiredToken)
				return
			}
		} else {
			app.jwtErrorResponse(w, r, utils.ErrInvalidClaims)
			return
		}

		userID, okID := claims["id"].(string)
		userRole, okRole := claims["userRole"].(string)

		if !okID || !okRole {
			app.jwtErrorResponse(w, r, utils.ErrInvalidClaims)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		ctx = context.WithValue(ctx, UserRoleKey, userRole)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (app *application) requireAdmin(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v := validator.New()
		if !app.isAdmin(v, r) {
			if len(v.Errors) > 0 {
				app.failedValidationResponse(w, r, v.Errors)
			} else {
				app.jwtErrorResponse(w, r, errors.New("you do not have permission to access this resource"))
			}
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Check if the user is an admin
func (app *application) isAdmin(v *validator.Validator, r *http.Request) bool {
	userIDStr, ok := r.Context().Value(UserRoleKey).(string)
	if !ok {
		v.AddError("Token", "User ID is missing from context")
		return false
	}
	userIDStrs, err := strconv.Atoi(userIDStr)
	if err != nil {
		return false
	}
	data.ValidatingUserRole(v, userIDStrs)
	return v.Valid()
}
func (app *application) requireVendorPermission(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vendorIDStr := r.PathValue("id")

		if vendorIDStr == "" {
			app.badRequestResponse(w, r, errors.New("vendor ID is required"))
			return
		}

		vendorID, err := uuid.Parse(vendorIDStr)
		if err != nil {
			app.badRequestResponse(w, r, errors.New("invalid vendor ID format"))
			return
		}

		v := validator.New()
		if app.isAdmin(v, r) {
			next.ServeHTTP(w, r)
			return
		}

		err = app.isVendorOwner(r, vendorID)
		if err != nil {
			if errors.Is(err, data.ErrRecordNotFound) {
				app.jwtErrorResponse(w, r, errors.New("you do not have permission to access this resource"))
			} else {
				app.jwtErrorResponse(w, r, errors.New("internal server error while checking vendor permissions"))
			}
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Check if the user is the owner of the vendor
func (app *application) isVendorOwner(r *http.Request, vendorID uuid.UUID) error {
	userIDStr, ok := r.Context().Value(UserIDKey).(string)
	if !ok {
		return errors.New("user ID is missing from context")
	}

	userid, err := uuid.Parse(userIDStr)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	ownerns, err := app.Model.VendorAdminDB.GetVendorAdmins(r.Context(), vendorID)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			return errors.New("you do not have permission to access this resource")
		}
		return errors.New("error checking vendor ownership")
	}
	for _, value := range ownerns {
		if value.UserID == userid {
			return nil // user is a vendor admin, so return no error
		}
	}
	return errors.New("you do not have permission to access this resource")
}
func (app *application) AuthorizeUserUpdate(next http.Handler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIDFromURL := r.PathValue("id")
		userIDFromContext, ok := r.Context().Value(UserIDKey).(string)
		if !ok {
			app.errorResponse(w, r, http.StatusUnauthorized, "user ID is missing from context")
			return
		}

		userRole, ok := r.Context().Value(UserRoleKey).(string)
		if !ok {
			app.errorResponse(w, r, http.StatusUnauthorized, "user role is missing from context")
			return
		}

		if r.Method == http.MethodPut {
			// Check if the user is updating their own account or is an admin
			if userIDFromContext != userIDFromURL && userRole != "1" {
				app.errorResponse(w, r, http.StatusForbidden, "you do not have permission to update this user")
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com")
		w.Header().Set("Referrer-Policy", "origin-when-cross-origin")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "deny")
		w.Header().Set("X-XSS-Protection", "0")

		// CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000") // Allow only your frontend's origin
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		// Handle preflight request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) logRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infoLog.Printf("%s - %s %s %s", r.RemoteAddr, r.Proto, r.Method,
			r.URL.RequestURI())
		next.ServeHTTP(w, r)
	})
}
func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()
		next.ServeHTTP(w, r)
	})
}

/*
func (app *application) rateLimit(next http.Handler) http.Handler {
	type client struct {
		limiter   *rate.Limiter
		lastSeen  time.Time
		banned    bool
		banExpiry time.Time
	}

	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(10 * time.Second) // Reset the limiter every 1 minute
			mu.Lock()
			for _, client := range clients {
				client.limiter = rate.NewLimiter(rate.Every(500*time.Millisecond), 10)
			}
			mu.Unlock()
		}
	}()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		mu.Lock()
		if _, found := clients[ip]; !found {
			clients[ip] = &client{limiter: rate.NewLimiter(rate.Every(10*time.Second), 30)} // Set the limit to 30 requests every 10 seconds
		}

		clients[ip].lastSeen = time.Now()

		if clients[ip].banned {
			if time.Now().Before(clients[ip].banExpiry) {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}
			clients[ip].banned = false
		}

		if !clients[ip].limiter.Allow() {
			clients[ip].banned = true
			clients[ip].banExpiry = time.Now().Add(30 * time.Second) // Ban for 1 minute
			mu.Unlock()
			app.rateLimitExceededResponse(w, r)
			return
		}

		mu.Unlock()
		next.ServeHTTP(w, r)
	})
}
*/
