package main

import (
	"errors"
	"net/http"
	"project/internal/data"
	"project/utils"
	"strconv"
	"time"

	"github.com/google/uuid"
)

func (app *application) GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	customerID, err := uuid.Parse(r.Context().Value(UserIDKey).(string))
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid customer ID"))
		return
	}
	// Retrieve the customer's table
	table, err := app.Model.TableDB.GetCustomertable(r.Context(), customerID)
	if err != nil {
		app.handleRetrievalError(w, r, err)
		return
	}
	orders, err := app.Model.OrderDB.GetOrders(customerID, table.ID)
	if err != nil {
		app.handleRetrievalError(w, r, err)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"orders": orders})
}
func (app *application) GetVendorOrdersHandler(w http.ResponseWriter, r *http.Request) {
	vendorID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid vendor ID"))
		return
	}

	orders, err := app.Model.OrderDB.GetVendorOrders(vendorID)
	if err != nil {
		app.handleRetrievalError(w, r, err)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"orders": orders})
}

// CreateOrderHandler handles the creation of a new order.
func (app *application) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		app.badRequestResponse(w, r, errors.New("failed to parse form"))
		return
	}

	totalOrderCostStr := r.FormValue("total_order_cost")
	customerIDStr := r.FormValue("customer_id")
	vendorIDStr := r.FormValue("vendor_id")
	status := r.FormValue("status")

	// Parse total_order_cost to a float64
	totalOrderCost, err := strconv.ParseFloat(totalOrderCostStr, 64)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid total order cost"))
		return
	}

	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid customer ID"))
		return
	}

	vendorID, err := uuid.Parse(vendorIDStr)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid vendor ID"))
		return
	}

	// Create a new order
	order := &data.Order{
		ID:             uuid.New(),
		TotalOrderCost: totalOrderCost,
		CustomerID:     customerID,
		VendorID:       vendorID,
		Status:         status,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Insert the order into the database
	err = app.Model.OrderDB.InsertOrder(order)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	utils.SendJSONResponse(w, http.StatusCreated, utils.Envelope{"order": order})
}

// DeleteOrderHandler handles the deletion of an order by its ID.
func (app *application) DeleteOrderHandler(w http.ResponseWriter, r *http.Request) {
	orderID, err := uuid.Parse(r.FormValue("id"))
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid order ID"))
		return
	}

	err = app.Model.OrderDB.DeleteOrder(orderID)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundResponse(w, r)
			return
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"message": "order deleted successfully"})
}
func (app *application) UpdateOrderStatusHandler(w http.ResponseWriter, r *http.Request) {
	orderIDStr := r.PathValue("id")
	status := r.FormValue("status")

	// Validate the order ID
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid order ID"))
		return
	}

	// Update the order status
	err = app.Model.OrderDB.UpdateOrder(orderID, status)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		if status == "completed" {
			// If status is completed, delete the order after updating
			err = app.Model.OrderDB.DeleteCompletedOrder(orderID)
			if err != nil {
				app.serverErrorResponse(w, r, err)
				return
			}
		}
	}

	// Respond with a success message
	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"message": "order status updated successfully"})
}
func (app *application) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(string)
	vendorID := r.PathValue("vendor_id")

	customerID, err := uuid.Parse(userID)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid user ID"))
		return
	}

	vendorUUID, err := uuid.Parse(vendorID)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid vendor ID"))
		return
	}

	// Retrieve the customer's table
	table, err := app.Model.TableDB.GetCustomertable(r.Context(), customerID)
	if err != nil {
		app.handleRetrievalError(w, r, err)
		return
	}

	// Retrieve the orders for the customer on the specific table
	orders, err := app.Model.OrderDB.GetOrders(customerID, table.ID)
	if err != nil {
		app.handleRetrievalError(w, r, err)
		return
	}

	// Filter orders by vendor
	var vendorOrders []data.OrderDetails
	for _, order := range orders {
		if order.VendorID == vendorUUID {
			vendorOrders = append(vendorOrders, order)
		}
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"orders": vendorOrders})
}
