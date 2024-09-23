package main

import (
	"errors"
	"net/http"
	"project/internal/data"
	"project/utils"
	"project/utils/validator"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (app *application) CreateItemHandler(w http.ResponseWriter, r *http.Request) {
	vendorID := r.PathValue("id")
	name := r.FormValue("name")
	priceStr := utils.NormalizeFloatInput(r.FormValue("price"))
	discountStr := utils.NormalizeFloatInput(r.FormValue("discount"))
	discountDaysStr := r.FormValue("discount_days")
	quantityStr := r.FormValue("quantity")

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid price"))
		return
	}

	discount, err := strconv.ParseFloat(discountStr, 64)
	if err != nil {
		discount = 0
	}

	quantity, err := strconv.Atoi(quantityStr)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid quantity"))
		return
	}

	var discountExpiresAt *time.Time
	if discountStr != "" && discount > 0 {
		if discountDaysStr == "" {
			app.badRequestResponse(w, r, errors.New("discount expiration date is required when a discount is provided"))
			return
		}

		discountDays, err := strconv.Atoi(discountDaysStr)
		if err != nil {
			app.badRequestResponse(w, r, errors.New("invalid discount days"))
			return
		}
		expiration := time.Now().Add(time.Duration(discountDays) * 24 * time.Hour)
		discountExpiresAt = &expiration
	}

	vendorsID, err := uuid.Parse(vendorID)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid vendor ID"))
		return
	}

	item := &data.Item{
		ID:             uuid.New(),
		VendorID:       vendorsID,
		Name:           name,
		Price:          price,
		Discount:       discount,
		DiscountExpiry: discountExpiresAt,
		Quantity:       quantity,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Validate item
	v := validator.New()
	data.ValidatingItem(v, item, "name", "price", "discount", "discount_expiry", "quantity")
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}

	if file, fileHeader, err := r.FormFile("img"); err == nil {
		defer file.Close()
		imageName, err := utils.SaveImageFile(file, "items", fileHeader.Filename)
		if err != nil {
			app.errorResponse(w, r, http.StatusBadRequest, "invalid image")
			return
		}
		item.Img = &imageName
	}

	err = app.Model.ItemDB.InsertItem(item)
	if err != nil {
		if item.Img != nil {
			utils.DeleteImageFile(*item.Img)
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	utils.SendJSONResponse(w, http.StatusCreated, utils.Envelope{"item": item})
}

// DeleteItemHandler handles the deletion of an item by its ID.
func (app *application) DeleteItemHandler(w http.ResponseWriter, r *http.Request) {
	itemID, err := uuid.Parse(r.PathValue("itemid"))
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid item ID"))
		return
	}

	err = app.Model.ItemDB.DeleteItem(itemID)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.errorResponse(w, r, http.StatusNotFound, "item already deleted")
		} else {
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"message": "item deleted successfully"})
}
func (app *application) GetAllItemsHandler(w http.ResponseWriter, r *http.Request) {
	vendorID := uuid.MustParse(r.PathValue("id"))

	// Create a Filters instance from request parameters
	var itemsSortSafelist = []string{"created_at", "name", "price"}
	filters := utils.Filters{
		Page:         1,                 // Default page
		PageSize:     10,                // Default page size
		Sort:         "created_at",      // Default sort
		SortSafelist: itemsSortSafelist, // Set the safelist
	}

	// Parse query parameters for filters
	if page := r.URL.Query().Get("page"); page != "" {
		filters.Page, _ = strconv.Atoi(page)
	}
	if pageSize := r.URL.Query().Get("page_size"); pageSize != "" {
		filters.PageSize, _ = strconv.Atoi(pageSize)
	}
	if sort := r.URL.Query().Get("sort"); sort != "" {
		filters.Sort = sort
	}
	if search := r.URL.Query().Get("search"); search != "" {
		filters.Search = search
	}

	v := validator.New()

	utils.ValidateFilters(v, filters)
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	items, err := app.Model.ItemDB.GetAllItems(vendorID, filters)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"items": items})
}
func (app *application) GetAllItemsCountHandler(w http.ResponseWriter, r *http.Request) {
	vendorID := uuid.MustParse(r.PathValue("id"))

	totalCount, err := app.Model.ItemDB.GetAllItemsCount(vendorID)
	if err != nil {
		app.handleRetrievalError(w, r, err)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{
		"totalCount": totalCount,
	})
}

// GetItemHandler handles fetching a single item by its ID.
func (app *application) GetItemHandler(w http.ResponseWriter, r *http.Request) {
	itemID, err := uuid.Parse(r.PathValue("itemid"))
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid item ID"))
		return
	}

	item, err := app.Model.ItemDB.GetItem(itemID)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundResponse(w, r)
		} else {
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"item": item})
}
func (app *application) UpdateItemHandler(w http.ResponseWriter, r *http.Request) {

	itemID, err := uuid.Parse(r.PathValue("itemid"))
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid item ID"))
		return
	}

	name := r.FormValue("name")
	priceStr := r.FormValue("price")
	discountStr := r.FormValue("discount")
	discountDaysStr := r.FormValue("discount_days")
	quantityStr := r.FormValue("quantity")

	price := utils.NormalizeFloatInput(priceStr)
	pricee, err := strconv.ParseFloat(price, 64)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid price"))
		return
	}
	// Get the existing item
	item, err := app.Model.ItemDB.GetItem(itemID)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundResponse(w, r)
		} else {
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	discount, err := strconv.ParseFloat(discountStr, 64)
	if err != nil && discountStr != "" {
		discount = 0 // Default to 0 if invalid
	}

	if name != "" {
		item.Name = name
	}
	if priceStr != "" {
		item.Price = pricee
	}
	if discountStr != "" {
		item.Discount = discount
		if discount > 0 {
			if discountDaysStr == "" {
				app.badRequestResponse(w, r, errors.New("discount expiration date is required when a discount is provided"))
				return
			}
			discountDays, err := strconv.Atoi(discountDaysStr)
			if err != nil {
				app.badRequestResponse(w, r, errors.New("invalid discount days"))
				return
			}
			expiration := time.Now().Add(time.Duration(discountDays) * 24 * time.Hour)
			item.DiscountExpiry = &expiration
		} else {
			item.DiscountExpiry = nil
		}
	}

	if quantityStr != "" {
		quantity, err := strconv.Atoi(quantityStr)
		if err != nil {
			app.badRequestResponse(w, r, errors.New("invalid quantity"))
			return
		}
		item.Quantity = quantity
	}

	v := validator.New()
	data.ValidatingItem(v, item, "name", "price", "discount", "discount_expiry", "quantity")
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
		return
	}
	var oldImg *string
	if item.Img != nil {
		*item.Img = strings.TrimPrefix(*item.Img, data.Domain+"/")
		oldImg = item.Img
	}

	if file, fileHeader, err := r.FormFile("img"); err == nil {
		defer file.Close()
		imageName, err := utils.SaveImageFile(file, "items", fileHeader.Filename)
		if err != nil {
			app.errorResponse(w, r, http.StatusBadRequest, "invalid image")
			return
		}
		if item.Img != nil {
			utils.DeleteImageFile(*item.Img)
		}
		item.Img = &imageName
	}

	err = app.Model.ItemDB.UpdateItem(item)
	if err != nil {
		if item.Img != nil {
			utils.DeleteImageFile(*item.Img)
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	if oldImg != nil && item.Img != nil && *oldImg != *item.Img {
		utils.DeleteImageFile(*oldImg)
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"item": item})
}
