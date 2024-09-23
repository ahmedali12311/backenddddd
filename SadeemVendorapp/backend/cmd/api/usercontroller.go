package main

import (
	"fmt"
	"net/http"
	"project/internal/data"
	"project/utils"
	"project/utils/validator"
	"strconv"
	"strings"

	"github.com/google/uuid"
	_ "github.com/joho/godotenv/autoload"
)

func (app *application) LoginHandler(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	v := validator.New()

	data.ValidatingUser(v, &data.User{Email: email, Password: password}, "email", "password")
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	user, err := app.Model.UserDB.GetUserByEmail(email)
	if err != nil {
		app.handleRetrievalError(w, r, err)
		return
	}

	if !utils.CheckPassword(user.Password, password) {
		app.notFoundResponse(w, r)
		return
	}
	users, err := app.Model.UserRoleDB.GetUserRole(user.ID)
	if err != nil {
		app.handleRetrievalError(w, r, err)
		return
	}

	userrole := strconv.Itoa(users.RoleID)

	token, err := utils.GenerateToken(user.ID.String(), userrole)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{
		"expires": "24 hours",
		"token":   token,
	})
}
func (app *application) IndexUserHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve query parameters for pagination, sorting, and searching
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")
	sortColumn := r.URL.Query().Get("sortColumn")
	sortDirection := r.URL.Query().Get("sortDirection")
	search := r.URL.Query().Get("search")

	// Set default values for page and pageSize
	page := 1
	pageSize := 10

	// Parse page and pageSize query parameters
	if pageStr != "" {
		parsedPage, err := strconv.Atoi(pageStr)
		if err != nil || parsedPage < 1 {
			page = 1
		} else {
			page = parsedPage
		}
	}

	if pageSizeStr != "" {
		parsedPageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil || parsedPageSize < 1 {
			pageSize = 10
		} else {
			pageSize = parsedPageSize
		}
	}

	// Validate sort direction
	if sortDirection != "" && sortDirection != "ASC" && sortDirection != "DESC" {
		sortDirection = "ASC" // Default sort direction
	}

	// Set default sort column to "created_at" if it's not provided or invalid
	validSortColumns := map[string]bool{
		"name":       true,
		"created_at": true,
		// Add other valid columns as needed
	}

	if !validSortColumns[sortColumn] {
		sortColumn = "created_at"
	}

	// Call the model method with the filters
	users, err := app.Model.UserDB.GetUsers(sortColumn, sortDirection, page, pageSize, search)
	if err != nil {
		app.handleRetrievalError(w, r, err)
		return
	}
	// Send JSON response
	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"users": users})
}

func (app *application) ShowUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	uuidcon, err := uuid.Parse(id)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	user, err := app.Model.UserDB.GetUser(uuidcon)
	if err != nil {
		app.errorResponse(w, r, http.StatusNotFound, fmt.Sprintf("User with %v was not found", uuidcon))
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"user": user})
}

func (app *application) SignupHandler(w http.ResponseWriter, r *http.Request) {
	v := validator.New()
	user := &data.User{
		Name:     r.FormValue("name"),
		Phone:    r.FormValue("phone"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
	}

	if user.Name == "" || user.Phone == "" || user.Email == "" || user.Password == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "Must fill all the fields")
		return
	}

	data.ValidatingUser(v, user, "name", "phone", "email", "password")
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	hashedPassword, err := utils.HashPassword(user.Password)
	if err != nil {
		app.errorResponse(w, r, http.StatusInternalServerError, "Error hashing password")
		return
	}
	user.Password = hashedPassword

	if file, fileHeader, err := r.FormFile("img"); err == nil {
		defer file.Close()
		imageName, err := utils.SaveImageFile(file, "users", fileHeader.Filename)
		if err != nil {
			app.errorResponse(w, r, http.StatusBadRequest, "invalid image ")
			return
		}
		user.Img = &imageName
	}

	if err = app.Model.UserDB.Insert(user); err != nil {
		if user.Img != nil {
			utils.DeleteImageFile(*user.Img)
		}
		app.handleRetrievalError(w, r, err)
		return
	}

	if _, err = app.Model.UserRoleDB.GrantRole(user.ID, 3); err != nil {
		app.badRequestResponse(w, r, err)
		return
	}

	utils.SendJSONResponse(w, http.StatusCreated, utils.Envelope{"user": user})
}

func (app *application) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	idint, err := uuid.Parse(id)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	user, err := app.Model.UserDB.GetUser(idint)
	if err != nil {
		app.errorResponse(w, r, http.StatusNotFound, "User not found")
		return
	}
	fmt.Println(user.Password)

	var oldImg *string
	if user.Img != nil {
		*user.Img = strings.TrimPrefix(*user.Img, data.Domain+"/")
		oldImg = user.Img
	}

	if name := r.FormValue("name"); name != "" {
		user.Name = name
	}

	if phone := r.FormValue("phone"); phone != "" {
		user.Phone = phone
	}

	if email := r.FormValue("email"); email != "" {
		user.Email = email
	}

	// Only update the password if a new password is provided
	if password := r.FormValue("password"); password != "" {
		hashedPassword, err := utils.HashPassword(password)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		user.Password = hashedPassword
	}

	if file, fileHeader, err := r.FormFile("img"); err == nil {
		defer file.Close()
		imageName, err := utils.SaveImageFile(file, "users", fileHeader.Filename)
		if err != nil {
			app.errorResponse(w, r, http.StatusInternalServerError, "Error saving image file: "+err.Error())
			return
		}
		user.Img = &imageName
	}

	v := validator.New()
	data.ValidatingUser(v, user, "name", "phone", "email", "password")
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	if err = app.Model.UserDB.Update(user); err != nil {
		if user.Img != nil && *user.Img != *oldImg {
			utils.DeleteImageFile(*user.Img)
		}
		app.handleRetrievalError(w, r, err)
		return
	}

	if oldImg != nil && user.Img != nil && *oldImg != *user.Img {
		utils.DeleteImageFile(*oldImg)
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{fmt.Sprintf("User %v", user.ID): "Updated successfully!"})
}

func (app *application) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	iduu, err := uuid.Parse(id)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	user, err := app.Model.UserDB.DeleteUser(iduu)
	if err != nil {
		app.handleRetrievalError(w, r, err)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"deleted user": user})
}
func (app *application) MeHandler(w http.ResponseWriter, r *http.Request) {

	uuiduser, err := uuid.Parse(r.Context().Value(UserIDKey).(string))
	if err != nil {

		app.errorResponse(w, r, http.StatusUnauthorized, "Invalid user uuid")

		return
	}
	user, err := app.Model.UserDB.GetUser(uuiduser)
	if err != nil {
		app.handleRetrievalError(w, r, err)
		return
	}

	userRole := r.Context().Value(UserRoleKey)

	// Create a response with user details and role
	response := map[string]interface{}{
		"user_info": user,
		"user_role": userRole,
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"me": response})
}
