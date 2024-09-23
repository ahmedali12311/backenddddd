package main

import (
	"errors"
	"net/http"
	"project/internal/data"
	"project/utils"

	"github.com/google/uuid"
)

// GetTablesHandler retrieves all tables from the database.
func (app *application) GetALLTablesHandler(w http.ResponseWriter, r *http.Request) {
	tables, err := app.Model.TableDB.GetTables(r.Context())
	if err != nil {
		app.handleRetrievalError(w, r, err)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"tables": tables})
}

// GetTablesHandler retrieves all tables for a specific vendor from the database.
func (app *application) GetTablesHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve vendor ID from URL path
	vendorIDStr := r.PathValue("id")
	if vendorIDStr == "" {
		app.notFoundResponse(w, r)
		return
	}

	vendorID, err := uuid.Parse(vendorIDStr)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid vendor ID"))
		return
	}

	// Get tables for the specific vendor
	tables, err := app.Model.TableDB.GetVendorTables(r.Context(), vendorID)
	if err != nil {
		if err == data.ErrRecordNotFound {
			http.Error(w, "No tables found for the specified vendor", http.StatusNotFound)
		} else {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"tables": tables})
}

// GetTableHandler retrieves a single table by its ID and ensures it belongs to the vendor specified in the URL.
func (app *application) GetTableHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve table ID from URL path
	tableIDStr := r.PathValue("table_id")
	if tableIDStr == "" {
		app.notFoundResponse(w, r)
		return
	}

	tableID, err := uuid.Parse(tableIDStr)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid table ID"))
		return
	}

	// Get the table from the database
	table, err := app.Model.TableDB.GetTable(r.Context(), tableID)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundResponse(w, r)
		} else {
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"table": table})
}

func (app *application) CreateTableHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve vendor ID from URL path
	vendorIDStr := r.PathValue("id")
	if vendorIDStr == "" {
		app.notFoundResponse(w, r)
		return
	}

	vendorID, err := uuid.Parse(vendorIDStr)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid vendor ID"))
		return
	}

	// Parse the form values
	err = r.ParseForm()
	if err != nil {
		app.badRequestResponse(w, r, errors.New("failed to parse form"))
		return
	}

	name := r.FormValue("name")
	isAvailableStr := r.FormValue("is_available")
	isNeedsServiceStr := r.FormValue("is_needs_service")

	// Parse boolean values
	isAvailable, err := utils.ParseBoolOrDefault(isAvailableStr, true)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid is_available value"))
		return
	}

	isNeedsService, err := utils.ParseBoolOrDefault(isNeedsServiceStr, false)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid is_needs_service value"))
		return
	}

	// Create a new table entity
	table := &data.Table{
		ID:              uuid.New(),
		Name:            name,
		VendorID:        vendorID,
		IsAvailable:     isAvailable,
		IsNeedsServices: isNeedsService,
	}
	tableCount, err := app.Model.TableDB.CountTablesForVendor(r.Context(), vendorID)
	if err != nil {

		app.serverErrorResponse(w, r, err)
		return
	}

	if tableCount >= 12 {
		app.badRequestResponse(w, r, errors.New("vendor has reached the maximum number of tables (12)"))
		return
	}

	// Insert the table into the database
	err = app.Model.TableDB.Insert(r.Context(), table)
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	utils.SendJSONResponse(w, http.StatusCreated, utils.Envelope{"table": table})
}

// DeleteTableHandler handles the deletion of a table by its ID.
func (app *application) DeleteTableHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve vendor ID from URL path
	tableIDStr := r.PathValue("table_id")
	if tableIDStr == "" {
		app.notFoundResponse(w, r)
		return
	}

	tableID, err := uuid.Parse(tableIDStr)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid Table ID"))
		return
	}

	// Get the table from the database
	_, err = app.Model.TableDB.GetTable(r.Context(), tableID)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundResponse(w, r)
		} else {
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	table, err := app.Model.TableDB.DeleteTable(r.Context(), tableID)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundResponse(w, r)
		} else {
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"message": "table deleted successfully", "table": table})
}

func (app *application) UpdateTableHandler(w http.ResponseWriter, r *http.Request) {
	tableIDStr := r.PathValue("table_id")
	if tableIDStr == "" {
		app.notFoundResponse(w, r)
		return
	}

	tableID, err := uuid.Parse(tableIDStr)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid vendor ID"))
		return
	}

	// Get the table from the database
	table, err := app.Model.TableDB.GetTable(r.Context(), tableID)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundResponse(w, r)
		} else {
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	if r.FormValue("name") != "" {
		table.Name = r.FormValue("name")
	}

	if err := app.Model.TableDB.Update(r.Context(), table); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"table": table})
}
func (app *application) UpdateTableNeedsServiceHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve table ID and customer ID from URL path
	tableIDStr := r.PathValue("table_id")
	customerIDStr := r.Context().Value(UserIDKey).(string)
	if tableIDStr == "" || customerIDStr == "" {
		app.notFoundResponse(w, r)
		return
	}

	tableID, err := uuid.Parse(tableIDStr)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid table ID"))
		return
	}

	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid customer ID"))
		return
	}

	// Get the table from the database
	table, err := app.Model.TableDB.GetTable(r.Context(), tableID)
	if err != nil {

		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundResponse(w, r)
		} else {

			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if table.CustomerID != nil && *table.CustomerID != uuid.Nil && *table.CustomerID != customerID {
		app.errorResponse(w, r, http.StatusConflict, "The table is not available! Try again later")
		return
	}

	isNeedsServiceStr := r.FormValue("is_needs_service")
	isNeedsService, err := utils.ParseBoolOrDefault(isNeedsServiceStr, false)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid is_needs_service value"))
		return
	}

	table.IsNeedsServices = isNeedsService
	err = app.Model.TableDB.AssignCustomer(r.Context(), tableID, customerID)
	if err != nil {
		app.handleRetrievalError(w, r, err)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"table": table})
}
func (app *application) FreeTableHandler(w http.ResponseWriter, r *http.Request) {
	tableIDStr := r.PathValue("table_id")
	customerIDStr := r.Context().Value(UserIDKey).(string) // Extract customer ID from context
	if tableIDStr == "" || customerIDStr == "" {
		app.notFoundResponse(w, r)
		return
	}

	// Parse the table ID and customer ID as UUIDs
	tableID, err := uuid.Parse(tableIDStr)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid table ID"))
		return
	}

	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid customer ID"))
		return
	}

	// Get the table from the database
	table, err := app.Model.TableDB.GetTable(r.Context(), tableID)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundResponse(w, r)
		} else {
			app.serverErrorResponse(w, r, err)
		}
		return
	}
	zeroUUID := uuid.UUID{}
	if table.CustomerID == nil || (table.CustomerID != nil && *table.CustomerID == zeroUUID) || *table.CustomerID != customerID {
		app.errorResponse(w, r, http.StatusConflict, "The table is not available! Try again later.")
		return
	}
	// Free the table
	if err := app.Model.TableDB.FreeTable(r.Context(), tableID); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Delete user orders
	if err := app.Model.OrderDB.DeleteUserOrders(customerID); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"message": "Table freed and user orders deleted successfully"})
}

func (app *application) FreeCustomerTableHandler(w http.ResponseWriter, r *http.Request) {
	// Retrieve table ID from URL path
	tableIDStr := r.PathValue("table_id")

	tableID, err := uuid.Parse(tableIDStr)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid table ID"))
		return
	}

	// Get the table from the database
	table, err := app.Model.TableDB.GetTable(r.Context(), tableID)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.notFoundResponse(w, r)
		} else {
			app.serverErrorResponse(w, r, err)
		}
		return
	}

	if err := app.Model.TableDB.FreeTableVendor(r.Context(), tableID, table.VendorID); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	// Delete user orders
	if err := app.Model.OrderDB.DeleteUserOrders(*table.CustomerID); err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"message": "Table freed and user orders deleted successfully"})
}

func (app *application) GetCustomertable(w http.ResponseWriter, r *http.Request) {

	customerIDStr := r.Context().Value(UserIDKey).(string) // Extract customer ID from context
	if customerIDStr == "" {
		app.notFoundResponse(w, r)
		return
	}

	customerID, err := uuid.Parse(customerIDStr)
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid customer ID"))
		return
	}

	table, err := app.Model.TableDB.GetCustomertable(r.Context(), customerID)
	if err != nil {
		if err != data.ErrUserHasNoTable {
			app.handleRetrievalError(w, r, err)
			return
		} else {
			utils.SendJSONResponse(w, http.StatusNotFound, utils.Envelope{"tables": "User has no tables"})
			return
		}

	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"tables": table})
}
