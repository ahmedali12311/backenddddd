package main

import (
	"errors"
	"fmt"
	"net/http"
	"project/internal/data"
	"project/utils"
	"strconv"

	"github.com/google/uuid"
)

// IndexUserRoles handles the listing of user roles
func (app *application) IndexUserRoles(w http.ResponseWriter, r *http.Request) {
	userRoles, err := app.Model.UserRoleDB.GetUserRoles()
	if err != nil {
		app.serverErrorResponse(w, r, err)
		return
	}
	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"user_roles": userRoles})
}

// ShowUserRoleHandler handles the retrieval of a user role by ID
func (app *application) ShowUserRoleHandler(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "Invalid ID")
		return
	}

	userRoles, err := app.Model.UserRoleDB.GetUserRole(id)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.errorResponse(w, r, http.StatusNotFound, fmt.Sprintf("User role with ID %v was not found", id))
			return
		}
		app.serverErrorResponse(w, r, err)
		return
	}
	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"user_roles": userRoles})
}

// UpdateUserRoleHandler handles the updating of a user's role
func (app *application) GrantRole(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "Invalid ID")
		return
	}

	newRole, err := strconv.Atoi(r.FormValue("role"))
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "Invalid role ID")
		return
	}

	if newRole == 2 {
		if r.FormValue("vendorID") == "" {
			app.errorResponse(w, r, http.StatusBadRequest, "Must enter the vendor ID")
			return
		}
		vendorID, err := uuid.Parse(r.FormValue("vendorID"))
		if err != nil {
			app.errorResponse(w, r, http.StatusBadRequest, err.Error())
			return
		}

		vendoradmin := data.VendorAdmin{
			UserID:   id,
			VendorID: vendorID,
		}
		_, err = app.Model.VendorAdminDB.InsertVendorAdmin(r.Context(), vendoradmin)
		if err != nil {
			app.handleRetrievalError(w, r, err)
			return
		}
	}

	user, err := app.Model.UserRoleDB.UpdateRole(id, newRole)
	if err != nil {
		if newRole == 2 {
			vendor, err := uuid.Parse(r.FormValue("vendorID"))
			if err != nil {
				app.errorResponse(w, r, http.StatusBadRequest, err)
				return
			}
			err = app.Model.VendorAdminDB.DeleteVendorAdmin(r.Context(), id, vendor)
			if err != nil {
				app.errorResponse(w, r, http.StatusBadRequest, err)
				return
			}

		}

		app.handleRetrievalError(w, r, err)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"Updated user role": user})
}

func (app *application) RevokeRoleHandler(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.FormValue("id"))
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "Invalid ID")
		return
	}

	role, err := strconv.Atoi(r.FormValue("user_role"))
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "Invalid role ID")
		return
	}

	err = app.Model.UserRoleDB.RevokeRole(id, role)
	if err != nil {
		if errors.Is(err, data.ErrRecordNotFound) {
			app.errorResponse(w, r, http.StatusNotFound, fmt.Sprintf("user with ID %v's role already deleted", id))
			return
		}

		app.handleRetrievalError(w, r, err)
		return
	}
	if role == 2 {
		_, err = app.Model.VendorDB.GetUserVendors(r.Context(), id)
		if err != nil {
			if err == data.ErrRecordNotFound {
				user, err := app.Model.UserRoleDB.UpdateRole(id, 3)
				fmt.Print(user)
				if err != nil {
					app.handleRetrievalError(w, r, err)
					return
				}

			} else {
				app.handleRetrievalError(w, r, err)
				return
			}
		}
	}
	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{fmt.Sprintf("Deleted user %v 's role ", id): role})

}
