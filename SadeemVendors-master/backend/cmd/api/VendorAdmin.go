package main

import (
	"errors"
	"fmt"
	"net/http"

	"project/internal/data"
	"project/utils"
	"project/utils/validator"

	"github.com/google/uuid"
)

func (app *application) GetVendorAdminHandler(w http.ResponseWriter, r *http.Request) {
	UserID := r.FormValue("user_id")
	if UserID == "" {
		app.errorResponse(w, r, http.StatusBadRequest, "invalid UserID")
		return
	}
	vendorIDUUID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "invalid vendor_id")
		return
	}

	userIDUUID, err := uuid.Parse(UserID)
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "invalid user_id")
		return
	}

	vendorAdmin, err := app.Model.VendorAdminDB.GetVendorAdmin(r.Context(), userIDUUID, vendorIDUUID)
	if err != nil {
		if err.Error() == "vendor admin not found" {
			app.errorResponse(w, r, http.StatusNotFound, err.Error())
			return
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"vendor_admin": vendorAdmin})
}
func (app *application) GetVendorAdminsHandler(w http.ResponseWriter, r *http.Request) {

	vendorIDUUID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "invalid vendor_id")
		return
	}

	vendorAdmin, err := app.Model.VendorAdminDB.GetVendorAdmins(r.Context(), vendorIDUUID)
	if err != nil {
		if err.Error() == "vendor admin not found" {
			app.errorResponse(w, r, http.StatusNotFound, err.Error())
			return
		}
		app.serverErrorResponse(w, r, err)
		return
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"vendor_admin": vendorAdmin})
}

func (app *application) DeleteVendorAdminHandler(w http.ResponseWriter, r *http.Request) {

	vendorIDUUID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "invalid vendor_id")
		return
	}

	UserID := r.PathValue("adminId")

	UserIDUUID, err := uuid.Parse(UserID)
	if err != nil {
		app.notFoundResponse(w, r)
		return
	}

	err = app.Model.VendorAdminDB.DeleteVendorAdmin(r.Context(), UserIDUUID, vendorIDUUID)
	if err != nil {
		switch {
		case errors.Is(err, data.ErrRecordNotFound):
			app.errorResponse(w, r, http.StatusNotFound, fmt.Sprintf("vendor admin with User ID %s and Vendor ID %s was not found", UserIDUUID, vendorIDUUID))
			return
		default:
			app.serverErrorResponse(w, r, err)
			return
		}
	}

	v, _ := app.Model.VendorDB.GetUserVendors(r.Context(), UserIDUUID)
	if v == nil {

		_, err := app.Model.UserRoleDB.UpdateRole(UserIDUUID, 3)
		if err != nil {
			app.handleRetrievalError(w, r, err)
			return

		}
	}

	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"message": "vendor admin deleted successfully"})
}

// CreateVendorAdminHandler handles the creation of a new vendor admin.
func (app *application) CreateVendorAdminHandler(w http.ResponseWriter, r *http.Request) {
	vendorIDUUID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "invalid vendor_id")
		return
	}

	UserEmail := r.FormValue("Email")

	user := &data.User{}
	user.Email = string(UserEmail)

	v := validator.New()

	data.ValidatingUser(v, user, "email")
	if !v.Valid() {
		app.failedValidationResponse(w, r, v.Errors)
	}

	getuser, err := app.Model.UserDB.GetUserByEmail(user.Email)
	if err != nil {
		app.handleRetrievalError(w, r, err)
		return
	}
	vendorAdmin := data.VendorAdmin{
		UserID:   getuser.ID,
		VendorID: vendorIDUUID,
	}
	_, err = app.Model.VendorDB.GetVendor(vendorIDUUID, true)
	if err != nil {

		app.handleRetrievalError(w, r, err)
		return
	}

	createdVendorAdmin, err := app.Model.VendorAdminDB.InsertVendorAdmin(r.Context(), vendorAdmin)
	if err != nil {

		app.handleRetrievalError(w, r, err)
		return
	}
	getuserrole, err := app.Model.UserRoleDB.GetUserRole(getuser.ID)
	if err != nil {
		app.handleRetrievalError(w, r, err)
		return
	}

	if getuserrole.RoleID == 3 {

		_, err = app.Model.UserRoleDB.UpdateRole(createdVendorAdmin.UserID, 2)
		if err != nil {
			if errors.Is(err, data.ErrDuplicatedRole) {

			} else {
				app.handleRetrievalError(w, r, err)

			}
			//don't shut down
		}
	}
	utils.SendJSONResponse(w, http.StatusCreated, utils.Envelope{"vendor_admin": createdVendorAdmin})
}

func (app *application) UpdateVendorAdminHandler(w http.ResponseWriter, r *http.Request) {
	vendorIDUUID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		app.errorResponse(w, r, http.StatusBadRequest, "invalid vendor_id")
		return
	}

	UserID, err := uuid.Parse(r.FormValue("User_ID"))
	if err != nil {
		app.badRequestResponse(w, r, errors.New("invalid user_id"))
		return
	}

	vendorAdmin := data.VendorAdmin{
		UserID:   UserID,
		VendorID: vendorIDUUID,
	}

	_, err = app.Model.VendorAdminDB.UpdateVendorAdmin(r.Context(), vendorAdmin)
	if err != nil {
		app.handleRetrievalError(w, r, err)
		return

	}
	utils.SendJSONResponse(w, http.StatusOK, utils.Envelope{"vendor_admin": vendorAdmin})
}
func (app *application) GetUserVendor(w http.ResponseWriter, r *http.Request) {
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
