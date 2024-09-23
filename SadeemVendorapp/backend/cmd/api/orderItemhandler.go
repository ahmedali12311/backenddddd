package main

import (
	"errors"
	"net/http"
	"project/internal/data"
	"project/utils"
	"strconv"

	"github.com/google/uuid"
)

// CreateOrderItemHandler handles the creation of a new order item.
func (app *application) CreateOrderItemHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.badRequestResponse(w, r, errors.New("failed to parse form"))
		return
	}

	orderIDStr := r.FormValue("order_id")
	itemIDStr := r.FormValue("item_id")
	quantityStr := r.FormValue("quantity")
	priceStr := r.FormValue("price")

	// Parse quantity and price to appropriate types
	quantity, err := strconv.Atoi(quantityStr)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid quantity"))
		return
	}

	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid price"))
		return
	}

	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid order ID"))
		return
	}

	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid item ID"))
		return
	}

	// Create a new order item
	orderItem := &data.OrderItem{
		ID:       uuid.New(),
		OrderID:  orderID,
		ItemID:   itemID,
		Quantity: quantity,
		Price:    price,
	}

	// Insert the order item into the database
	err = app.Model.OrderItemDB.InsertOrderItem(orderItem)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	utils.SendJSONResponse(w, http.StatusCreated, utils.Envelope{"order_item": orderItem})
}

// DeleteOrderItemHandler handles the deletion of an order item by its ID.
func (app *application) DeleteOrderItemHandler(w http.ResponseWriter, r *http.Request) {
	orderItemIDStr := r.FormValue("id")

	orderItemID, err := uuid.Parse(orderItemIDStr)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid order item ID"))
		return
	}

	err = app.Model.OrderItemDB.DeleteOrderItem(orderItemID)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundResponse(w, r)
		} else {
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"message": "order item deleted successfully"})
}
