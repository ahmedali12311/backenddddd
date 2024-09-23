package main

import (
	"errors"
	"net/http"
	"project/internal/data"
	"project/utils"
	"strings"
	"time"

	"github.com/google/uuid"
)

// CreateCartHandler handles the creation of a new cart.
func (app *application) CreateCartHandler(w http.ResponseWriter, r *http.Request) {

	err := app.Model.CartDB.InsertCart(&data.Cart{ID: uuid.MustParse(r.Context().Value(UserIDKey).(string))})
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	utils.SendJSONResponse(w, http.StatusCreated, utils.Envelope{"created cart for user ": r.Context().Value(UserIDKey).(string)})
}

// DeleteCartHandler handles the deletion of a cart by its ID.
func (app *application) DeleteCartHandler(w http.ResponseWriter, r *http.Request) {
	cartID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid cart ID"))
		return
	}

	err = app.Model.CartDB.DeleteCart(cartID)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundResponse(w, r)
		} else {
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"message": "cart deleted successfully"})
}
func (app *application) UpdateCartHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context
	userID := uuid.MustParse(r.Context().Value(UserIDKey).(string))

	// Get cart ID from request
	cartIDStr := r.PathValue("id")
	cartID, err := uuid.Parse(cartIDStr)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid cart ID"))
		return
	}

	// Fetch the current cart
	cart, err := app.Model.CartDB.GetCart(userID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Fetch all cart items to calculate the new total price
	cartItems, err := app.Model.CartItemDB.GetCartItems(cartID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Calculate the new total price
	totalPrice := 0.0
	for _, item := range cartItems {
		itemPrice, err := app.Model.ItemDB.GetItemPrice(item.ItemID) // Fetch the item price
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		if itemPrice.Discount != 0 {
			totalPrice += itemPrice.Discount * float64(item.Quantity) // Use the discount price
		} else {
			totalPrice += itemPrice.Price * float64(item.Quantity) // Use the regular price
		}
	}

	// Update the cart's total price and quantity
	cart.TotalPrice = totalPrice
	cart.Quantity = len(cartItems) // Update quantity based on the number of items
	cart.VendorID = uuid.MustParse(r.FormValue("vendor_id"))
	// Update the cart in the database
	err = app.Model.CartDB.UpdateCart(cart)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"cart": cart})
}

func (app *application) GetCartHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from the context (assuming you have middleware that sets this)
	userID := uuid.MustParse(r.Context().Value(UserIDKey).(string))

	// Retrieve all items in the cart
	cartItems, err := app.Model.CartDB.GetCart(userID)
	if err != nil {
		utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"cart": nil})
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"cart": cartItems})
}
func (app *application) CheckoutHandler(w http.ResponseWriter, r *http.Request) {

	userID := uuid.MustParse(r.Context().Value(UserIDKey).(string))
	_, err := app.Model.TableDB.GetCustomertable(r.Context(), userID)
	if err != nil {
		app.handleRetrievalError(w, r, err)
		return
	}
	// Fetch the cart for the user
	cart, err := app.Model.CartDB.GetCart(userID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Fetch all cart items
	cartItems, err := app.Model.CartItemDB.GetCartItems(cart.ID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Check vendor consistency
	if len(cartItems) == 0 {
		app.errorResponse(w, r, http.StatusBadRequest, "Cart is empty.")
		return
	}

	// Use the cart's VendorID for validation
	vendorID := cart.VendorID
	for _, item := range cartItems {
		itemData, err := app.Model.ItemDB.GetItem(item.ItemID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		if itemData.VendorID != vendorID {
			app.errorResponse(w, r, http.StatusBadRequest, "All items in the cart must be from the same vendor.")
			return
		}
	}

	// Begin a transaction
	tx, err := app.Model.BeginTransaction()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback() // Rollback on error
		} else {
			err = tx.Commit() // Commit if no errors
		}
	}()

	// Create a new order
	order := &data.Order{
		ID:             uuid.New(),
		TotalOrderCost: cart.TotalPrice,
		CustomerID:     userID,
		VendorID:       vendorID,
		Status:         "preparing",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Insert the order into the database
	if err = tx.InsertOrder(order); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Insert order items
	for _, item := range cartItems {
		// Fetch the item to get its price
		itemData, err := app.Model.ItemDB.GetItem(item.ItemID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		orderItem := &data.OrderItem{
			ID:       uuid.New(),
			OrderID:  order.ID,
			ItemID:   item.ItemID,
			Quantity: item.Quantity,
			Price:    itemData.Price,
		}
		if err = tx.InsertOrderItem(orderItem); err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		newQuantity := itemData.Quantity - item.Quantity
		if newQuantity < 0 {
			app.serverErrorResponse(w, r, errors.New("not enough quantity"))
			return
		}
		itemData.Quantity = newQuantity
		if itemData.Img != nil {
			*itemData.Img = strings.TrimPrefix(*itemData.Img, data.Domain+"/")
		}
		if err = app.Model.ItemDB.UpdateItem(itemData); err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	// Delete the cart and cart items
	if err = tx.DeleteCart(cart.ID); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if err = tx.DeleteCartItems(cart.ID); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// If we reach here, commit the transaction
	if err != nil {
		tx.Rollback() // Rollback on error
	} else {
		err = tx.Commit() // Commit if no errors
	}

	// Respond with a success message
	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"message": "checkout successful", "order_id": order.ID})
}
