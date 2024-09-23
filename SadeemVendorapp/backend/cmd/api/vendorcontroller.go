package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"project/internal/data"
	"project/utils"
	"project/utils/validator"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

func (app *application) IndexVendorHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve query parameters for pagination
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")
	sort := r.URL.Query().Get("sort")
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

	// Set default sort to "latest" if it's not provided or invalid
	if sort == "" || !validator.In(sort, "latest", "name_asc", "name_desc") {
		sort = "latest"
	}

	filters := utils.Filters{
		Page:         page,
		PageSize:     pageSize,
		Sort:         sort,
		SortSafelist: []string{"latest", "name_asc", "name_desc"},
		Search:       search,
	}

	v := validator.New()
	utils.ValidateFilters(v, filters)

	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Retrieve user role from context
	isAdmin, ok := r.Context().Value(UserRoleKey).(string)
	if !ok {
		isAdmin = ""
	}

	var vendors []data.Vendor
	var count int
	var err error

	// Handle the user ID from context
	userIDStr, _ := r.Context().Value(UserIDKey).(string)
	userID, _ := uuid.Parse(userIDStr)

	if isAdmin == "2" {
		// Vendor owner is also considered admin in this case
		vendors, err = app.Model.VendorDB.GetUserVendors(r.Context(), userID)

		if err != nil {
			app.handleRetrievalError(w, r, err)
			return
		}
		count = len(vendors) // Set count to the number of retrieved vendors

	} else {
		isVisible := isAdmin == "1" // Only admins (role 1) see all vendors; others see only visible ones

		// Fetch vendors with pagination
		vendorsPtr, totalCount, err := app.Model.VendorDB.GetVendors(filters, isVisible)
		if err != nil {
			app.handleRetrievalError(w, r, err)
			return
		}
		count = totalCount
		vendors = *vendorsPtr
	}

	// Prepare response
	response := utils.Envelope{
		"Vendors":    vendors,
		"TotalCount": count,
		"Page":       page,
		"PageSize":   pageSize,
	}

	utils.SendJSONResponse(w, http.StatusOK, response)
}

func (app *application) CreateVendor(w http.ResponseWriter, r *http.Request) {
	var vendor data.Vendor
	var newImage *string

	// Parse form values
	vendor.Name = r.FormValue("name")
	vendor.Description = r.FormValue("description")

	// Handle file upload
	file, fileHeader, err := r.FormFile("img")
	if err != nil && err != http.ErrMissingFile {
		app.badRequestResponse(w, r, errors.New("invalid file"))
		return
	} else if err == nil {
		defer file.Close()
		imageName, err := utils.SaveImageFile(file, "vendors", fileHeader.Filename)
		if err != nil {
			app.errorResponse(w, r, http.StatusInternalServerError, "Error saving image")
			return
		}
		vendor.Img = &imageName
		newImage = &imageName
	}

	if r.FormValue("subscriptionDays") != "" {
		// Parse subscription days
		subscriptionDaysStr := r.FormValue("subscriptionDays")
		subscriptionDays, err := strconv.Atoi(subscriptionDaysStr)
		if err != nil {
			app.errorResponse(w, r, http.StatusBadRequest, " subscription days error")

		}
		vendor.SubscriptionDays = subscriptionDays
	} else {
		vendor.SubscriptionDays = 30
	}

	// Validate the vendor data
	v := validator.New()
	data.ValidatingVendor(v, &vendor)
	if !v.Valid() {
		if newImage != nil {
			utils.DeleteImageFile(*newImage)
		}
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	// Insert the vendor into the database
	err = app.Model.VendorDB.InsertVendor(&vendor)
	if err != nil {
		if newImage != nil {
			utils.DeleteImageFile(*newImage)
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	// Send the response
	utils.SendJSONResponse(w, http.StatusCreated, utils.Envelope{"vendor created successfully ": vendor.ID})
}
func (app *application) UpdateVendorHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Retrieve the existing vendor
	vendor, err := app.Model.VendorDB.GetVendor(id, true) // Adjusted to handle three return values
	if err != nil {
		app.handleRetrievalError(w, r, err)
		return
	}
	if vendor == nil {
		app.notFoundResponse(w, r)
		return
	}

	var oldImg *string
	if vendor.Img != nil {
		*vendor.Img = strings.TrimPrefix(*vendor.Img, data.Domain+"/")
		oldImg = vendor.Img
	}

	if r.FormValue("name") != "" {
		vendor.Name = r.FormValue("name")
	}
	if r.FormValue("description") != "" {
		vendor.Description = r.FormValue("description")
	}
	if r.FormValue("subscriptionDays") != "" {
		vendor.SubscriptionDays, err = strconv.Atoi(r.FormValue("subscriptionDays"))
		if err != nil {
			app.errorResponse(w, r, http.StatusBadRequest, "Invalid subscription days")
			return
		}
	}

	if file, fileHeader, err := r.FormFile("img"); err == nil {
		defer file.Close()
		imageName, err := utils.SaveImageFile(file, "users", fileHeader.Filename)
		if err != nil {
			app.errorResponse(w, r, http.StatusInternalServerError, "Error saving image file: "+err.Error())
			return
		}
		vendor.Img = &imageName
	}

	v := validator.New()
	data.ValidatingVendor(v, vendor)
	if !v.Valid() {
		if oldImg != nil && vendor.Img != nil && *oldImg != *vendor.Img {
			utils.DeleteImageFile(*vendor.Img)
		}
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	err = app.Model.VendorDB.UpdateVendor(vendor)
	if err != nil {
		if vendor.Img != nil && oldImg != nil {
			utils.DeleteImageFile(*vendor.Img)
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	if oldImg != nil && vendor.Img != nil && *oldImg != *vendor.Img {
		utils.DeleteImageFile(*oldImg)
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"vendor": vendor})
}

func (app *application) DeleteVendorHandler(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	iduu, err := uuid.Parse(id)
	if err != nil {
		app.notFoundResponse(w, r)
		return // Ensure to return after handling the error
	}
	vendor, err := app.Model.VendorDB.DeleteVendor(iduu)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.errorResponse(w, r, http.StatusNotFound, fmt.Sprintf("vendor with ID %s was not found", id))
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}
	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"deleted vendor": vendor})
}
func (app *application) GetUserVendors(w http.ResponseWriter, r *http.Request) {
	userUUID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "invalid vendor_id")
		return
	}
	vendor, err := app.Model.VendorDB.GetUserVendors(r.Context(), userUUID)
	if err != nil {
		app.handleRetrievalError(w, r, err)
		return
	}
	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"vendor": vendor})
}
func (app *application) ShowVendorHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	// Retrieve user role from context
	isAdminRole, ok := r.Context().Value(UserRoleKey).(string)
	isAdmin := ok && (isAdminRole == "1" || isAdminRole == "2")

	vendor, err := app.Model.VendorDB.GetVendor(id, isAdmin)
	if err != nil {

		switch {
		case errors.Is(err, sql.ErrNoRows):
			app.errorResponse(w, r, http.StatusNotFound, fmt.Sprintf("Vendor with ID %s was not found", id))
			return
		default:
			app.badRequestResponse(w, r, err)
			return
		}
	}

	// Check if the vendor is visible, unless the user is an admin
	if !isAdmin && !vendor.IsVisible {
		app.errorResponse(w, r, http.StatusNotFound, "Vendor not found or not visible")
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"vendor": vendor})
}
func (app *application) GetVendorTablesHandler(w http.ResponseWriter, r *http.Request) {
	vendorIDs := r.PathValue("id")
	id, err := uuid.Parse(vendorIDs)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	// Get vendor tables from the database
	tables, err := app.Model.TableDB.GetVendorTables(r.Context(), id)
	if err != nil {
		app.handleRetrievalError(w, r, err)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"tables": tables})
}
