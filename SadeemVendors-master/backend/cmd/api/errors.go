package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"project/internal/data"
	"project/utils"
)

func (app *application) handleRetrievalError(w http.ResponseWriter, r *http.Request, err error) {
	switch {
	case errors.Is(err, data.ErrRecordNotFound):
		app.errorResponse(w, r, http.StatusNotFound, "Resources could not be found")
	case errors.Is(err, data.ErrDuplicatedKey):
		app.errorResponse(w, r, http.StatusConflict, "Email already exists, try something else")
	case errors.Is(err, data.ErrUserNotFound):
		app.errorResponse(w, r, http.StatusConflict, "Email not found, try again")
	case errors.Is(err, data.ErrDuplicatedRole):
		app.errorResponse(w, r, http.StatusConflict, "User already have the role")
	case errors.Is(err, data.ErrHasRole):
		app.errorResponse(w, r, http.StatusConflict, "User already has a role")
	case errors.Is(err, data.ErrHasNoRoles):
		app.errorResponse(w, r, http.StatusConflict, data.ErrHasNoRoles.Error())
	case errors.Is(err, data.ErrForeignKeyViolation):
		app.errorResponse(w, r, http.StatusNotFound, "Vendor ID is incorrect")
	case errors.Is(err, data.ErrUserAlreadyhaveatable):
		app.errorResponse(w, r, http.StatusConflict, data.ErrUserAlreadyhaveatable.Error())
	case errors.Is(err, data.ErrUserHasNoTable):
		app.errorResponse(w, r, http.StatusNotFound, data.ErrUserHasNoTable.Error())
	case errors.Is(err, data.ErrItemAlreadyInserted):
		app.errorResponse(w, r, http.StatusConflict, data.ErrItemAlreadyInserted.Error())
	case errors.Is(err, data.ErrInvalidQuantity):
		app.errorResponse(w, r, http.StatusConflict, data.ErrInvalidQuantity.Error())
	default:
		app.serverErrorResponse(w, r, err)
	}
}
func (app *application) errorResponse(w http.ResponseWriter, r *http.Request, status int, message interface{}) {
	env := utils.Envelope{"error": message}
	err := utils.SendJSONResponse(w, status, env)
	if err != nil {
		app.logError(r, err)
		w.WriteHeader(500)

	}
}

func (app *application) logError(r *http.Request, err error) {
	log.Printf("Error: %v, Method: %s, URL: %s", err, r.Method, r.URL.String())
}
func (app *application) serverErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.logError(r, err)
	message := "the server encountered a problem and could not process your request"
	app.errorResponse(w, r, http.StatusInternalServerError, message)
}
func (app *application) notFoundResponse(w http.ResponseWriter, r *http.Request) {
	message := "resources not found"
	app.errorResponse(w, r, http.StatusNotFound, message)
}
func (app *application) badRequestResponse(w http.ResponseWriter, r *http.Request, err error) {
	app.errorResponse(w, r, http.StatusBadRequest, err.Error())
}
func (app *application) failedValidationResponse(w http.ResponseWriter, r *http.Request, errors map[string]string) {
	app.errorResponse(w, r, http.StatusUnprocessableEntity, errors)
}
func (app *application) jwtErrorResponse(w http.ResponseWriter, r *http.Request, err error) {
	var message string
	switch {
	case errors.Is(err, utils.ErrInvalidToken):
		message = "invalid token"
	case errors.Is(err, utils.ErrExpiredToken):
		message = "token has expired"
	case errors.Is(err, utils.ErrMissingToken):
		message = "missing authorization token"
	case errors.Is(err, utils.ErrInvalidClaims):
		message = "invalid token claims"
	default:
		app.errorResponse(w, r, http.StatusUnauthorized, "You don't have a premission")
		return
	}
	app.errorResponse(w, r, http.StatusUnauthorized, message)
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func (app *application) ErrorHandlerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the error
				app.log.Println("Recovered from panic:", err)

				// Send the error response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)

				response := ErrorResponse{
					Error: "Internal Server Error",
				}
				json.NewEncoder(w).Encode(response)
			}
		}()

		next.ServeHTTP(w, r)
	})
}

/*
func (app *application) rateLimitExceededResponse(w http.ResponseWriter, r *http.Request) {
	message := "rate limit exceeded"
	app.errorResponse(w, r, http.StatusTooManyRequests, message)
}
*/
