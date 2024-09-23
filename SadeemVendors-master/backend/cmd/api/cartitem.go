package main

import (
	"errors"
	"net/http"
	"project/internal/data"
	"project/utils"
	"strconv"

	"github.com/google/uuid"
)

func (app *application) CreateCartItemHandler(w http.ResponseWriter, r *http.Request) {
	cartIDStr := r.Context().Value(UserIDKey).(string)
	itemIDStr := r.FormValue("item_id")
	quantityStr := r.FormValue("quantity")

	quantity, err := strconv.Atoi(quantityStr)
	if err != nil || quantity <= 0 {
		app.badRequestResponse(w, r, errors.New("invalid quantity"))
		return
	}

	cartID, err := uuid.Parse(cartIDStr)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid cart ID"))
		return
	}

	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid item ID"))
		return
	}

	// Check if item is available in the required quantity
	isAvailable, err := app.Model.ItemDB.IsStockAvailable(itemID, quantity)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if !isAvailable {
		app.handleRetrievalError(w, r, err)
		return
	}

	_, err = app.Model.ItemDB.GetItem(itemID)
	if err != nil {
		app.handleRetrievalError(w, r, err)
		return
	}

	itemVendorID, err := app.Model.ItemDB.GetVendorID(itemID)
	if err != nil {
		app.handleRetrievalError(w, r, err)
		return
	}
	cart, err := app.Model.CartDB.GetCart(cartID)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			err = app.Model.CartDB.InsertCart(&data.Cart{
				ID:       uuid.MustParse(r.Context().Value(UserIDKey).(string)),
				VendorID: itemVendorID,
			})
			if err != nil {
				app.handleRetrievalError(w, r, err)
				return
			}
			cart, err = app.Model.CartDB.GetCart(cartID)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}
		} else {
			app.notFoundResponse(w, r)
			return
		}
	}
	// Check if the user has a table
	_, err = app.Model.TableDB.GetCustomertable(r.Context(), cartID)
	if err != nil {
		app.errorResponse(w, r, http.StatusForbidden, "You must have a table to add items.")
		return
	}
	// Check vendor consistency
	if cart.VendorID != uuid.Nil && cart.VendorID != itemVendorID {
		app.errorResponse(w, r, http.StatusBadRequest, "You can only add items from the same vendor to this cart.")
		return
	}

	// Create a new cart item
	cartItem := &data.CartItem{
		CartID:   cartID,
		ItemID:   itemID,
		Quantity: quantity,
	}

	// Fetch item price
	itemPrice, err := app.Model.ItemDB.GetItemPrice(itemID)
	if err != nil {
		app.handleRetrievalError(w, r, err)
		return
	}

	// Insert the cart item
	err = app.Model.CartItemDB.InsertCartItem(cartItem)
	if err != nil {
		app.handleRetrievalError(w, r, err)
		return
	}

	// Update total price and quantity in cart
	cart, err = app.Model.CartDB.GetCart(cartID)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}
	if itemPrice.Discount == 0 {
		cart.Quantity += quantity
		cart.TotalPrice += itemPrice.Price * float64(quantity)
		// Update the cart in the database
	} else {
		cart.Quantity += quantity
		cart.TotalPrice += itemPrice.Discount * float64(quantity)
	}
	err = app.Model.CartDB.UpdateCart(cart)
	if err != nil {
		app.handleRetrievalError(w, r, err)
		return
	}

	// Respond with the created cart item
	utils.SendJSONResponse(w, http.StatusCreated, utils.Envelope{"cart_item": cartItem})
}
func (app *application) DeleteCartItemHandler(w http.ResponseWriter, r *http.Request) {
	cartIDStr := r.Context().Value(UserIDKey).(string)
	itemIDStr := r.PathValue("id")
	quantityStr := r.FormValue("quantity")

	cartID, err := uuid.Parse(cartIDStr)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid cart ID"))
		return
	}

	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid item ID"))
		return
	}

	// Parse the quantity string to an integer
	quantity, err := strconv.Atoi(quantityStr)
	if err != nil {
		quantity = 0
	}

	// Fetch the current item to get its price and quantity
	cartItems, err := app.Model.CartItemDB.GetCartItems(cartID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	var currentItem *data.CartItem
	for _, item := range cartItems {
		if item.ItemID == itemID {
			currentItem = &item
			break
		}
	}

	if currentItem == nil {
		app.notFoundResponse(w, r)
		return
	}

	// Get the item's price to adjust the cart's total price
	itemPrice, err := app.Model.ItemDB.GetItemPrice(itemID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Fetch the cart for updating
	cart, err := app.Model.CartDB.GetCart(cartID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	if quantity == 0 {
		// Delete the entire cart item
		err = app.Model.CartItemDB.DeleteCartItem(r.Context(), cartID, itemID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		// Update the cart's total price and quantity
		if itemPrice.Discount == 0 {
			cart.TotalPrice -= itemPrice.Price * float64(currentItem.Quantity)
			cart.Quantity -= currentItem.Quantity
		} else {
			cart.TotalPrice -= itemPrice.Discount * float64(currentItem.Quantity)
			cart.Quantity -= currentItem.Quantity
		}
	} else if quantity < 0 {
		// Return an error message if quantity is less than 0
		app.badRequestResponse(w, r, errors.New("quantity must be greater than 0"))
		return
	} else {
		if quantity > currentItem.Quantity {
			app.errorResponse(w, r, http.StatusBadRequest, "cannot remove more items than exist in the cart")
			return
		}

		// Adjust total price by removing the item's total cost
		cart.TotalPrice -= itemPrice.Price * float64(quantity)
		cart.Quantity -= quantity

		// Update the cart item quantity
		currentItem.Quantity -= quantity
		err = app.Model.CartItemDB.Updatecartitem(currentItem)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	if cart.Quantity == 0 {
		cart.VendorID = uuid.Nil
		err = app.Model.CartDB.DeleteCart(cartID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	} else {
		// Update the cart in the database
		err = app.Model.CartDB.UpdateCart(cart)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	// Respond with a success message
	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"message": "cart item deleted successfully"})
}

func (app *application) UpdateCartItemHandler(w http.ResponseWriter, r *http.Request) {
	cartIDStr := r.Context().Value(UserIDKey).(string)
	itemIDStr := r.FormValue("item_id")
	quantityStr := r.FormValue("quantity")

	quantity, err := strconv.Atoi(quantityStr)
	if err != nil || quantity < 0 {
		app.badRequestResponse(w, r, errors.New("invalid quantity"))
		return
	}

	cartID, err := uuid.Parse(cartIDStr)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid cart ID"))
		return
	}

	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid item ID"))
		return
	}

	// Fetch current items in the cart
	cartItems, err := app.Model.CartItemDB.GetCartItems(cartID)
	if err != nil {
		app.handleRetrievalError(w, r, err)
		return
	}

	var currentItem *data.CartItem
	for _, item := range cartItems {
		if item.ItemID == itemID {
			currentItem = &item
			break
		}
	}

	if currentItem == nil {
		app.notFoundResponse(w, r)
		return
	}

	// Check stock availability for the new quantity
	isAvailable, err := app.Model.ItemDB.IsStockAvailable(itemID, quantity)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	if !isAvailable {
		app.errorResponse(w, r, http.StatusConflict, "requested quantity is not available")
		return
	}

	// Fetch item price and vendor
	itemPrice, err := app.Model.ItemDB.GetItemPrice(itemID)
	if err != nil {
		app.handleRetrievalError(w, r, err)
		return
	}

	itemVendorID, err := app.Model.ItemDB.GetVendorID(itemID)
	if err != nil {
		app.handleRetrievalError(w, r, err)
		return
	}

	// Fetch the cart for updating
	cart, err := app.Model.CartDB.GetCart(cartID)
	if err != nil {
		app.handleRetrievalError(w, r, err)
		return
	}

	// Check vendor consistency
	if cart.VendorID != uuid.Nil && cart.VendorID != itemVendorID {
		app.errorResponse(w, r, http.StatusBadRequest, "You can only update items from the same vendor in this cart.")
		return
	}

	// Calculate the difference in quantity
	difference := quantity - currentItem.Quantity

	if difference > 0 {
		// Increasing quantity
		if itemPrice.Discount != 0 {
			cart.TotalPrice += float64(difference) * itemPrice.Price
			cart.Quantity += difference
		} else {
			cart.TotalPrice += float64(difference) * itemPrice.Discount
			cart.Quantity += difference
		}
	} else {
		// Decreasing quantity
		if itemPrice.Discount == 0 {
			cart.TotalPrice -= float64(-difference) * itemPrice.Price
			cart.Quantity -= -difference
		} else {
			cart.TotalPrice -= float64(-difference) * itemPrice.Discount
			cart.Quantity -= -difference
		}
	}
	// Update the current item quantity
	currentItem.Quantity = quantity
	err = app.Model.CartItemDB.Updatecartitem(currentItem)
	if err != nil {
		app.handleRetrievalError(w, r, err)
		return
	}

	// Update the cart in the database
	err = app.Model.CartDB.UpdateCart(cart)
	if err != nil {
		app.handleRetrievalError(w, r, err)
		return
	}

	// Respond with the updated cart item
	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"cart_item": currentItem})
}

func (app *application) GetCartItemswithimage(w http.ResponseWriter, r *http.Request) {
	// Get user ID from the context (assuming you have middleware that sets this)
	cartID := uuid.MustParse(r.Context().Value(UserIDKey).(string))

	// Retrieve all items in the cart
	cartItems, err := app.Model.CartItemDB.GetCartItemswithimage(cartID)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"cart": cartItems})
}
