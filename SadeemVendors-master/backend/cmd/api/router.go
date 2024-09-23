package main

import (
	"net/http"

	"github.com/go-michi/michi"
)

func (app *application) Router() *michi.Router {
	r := michi.NewRouter()
	// Apply global middleware
	r.Use(app.logRequest)
	r.Use(app.recoverPanic)
	r.Use(secureHeaders)
	r.Use(app.ErrorHandlerMiddleware)
	/* 	r.Use(app.rateLimit)
	 */
	r.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	r.Route("/", func(sub *michi.Router) {
		// User routes
		sub.HandleFunc("GET users", app.AuthMiddleware(http.HandlerFunc(app.requireAdmin(http.HandlerFunc(app.IndexUserHandler)))))
		sub.HandleFunc("GET users/{id}", app.AuthMiddleware(http.HandlerFunc(app.ShowUserHandler)))
		sub.HandleFunc("PUT users/{id}", app.AuthMiddleware(http.HandlerFunc(app.AuthorizeUserUpdate(http.HandlerFunc(app.UpdateUserHandler)))))
		sub.HandleFunc("DELETE users/{id}", app.AuthMiddleware(http.HandlerFunc(app.requireAdmin(http.HandlerFunc(app.DeleteUserHandler)))))
		// Auth routes (public)
		sub.HandleFunc("POST signin", http.HandlerFunc(app.LoginHandler))
		sub.HandleFunc("POST signup", http.HandlerFunc(app.SignupHandler))
		// Table routes
		//to get the table details of assigned customer's table
		sub.HandleFunc("GET usertable", app.AuthMiddleware(http.HandlerFunc(app.GetCustomertable)))
		//to get the table details of vendor's tables
		sub.HandleFunc("GET  vendor/{id}/tables", app.AuthMiddleware(http.HandlerFunc(app.GetTablesHandler)))
		//to get the table details of vendor's table
		sub.HandleFunc("GET vendor/{id}/tables/{table_id}", app.AuthMiddleware(http.HandlerFunc(app.requireVendorPermission(http.HandlerFunc(app.GetTableHandler)))))
		//to add the table of a vendor
		sub.HandleFunc("POST vendor/{id}/tables", app.AuthMiddleware(http.HandlerFunc(app.requireVendorPermission(http.HandlerFunc(app.CreateTableHandler)))))
		//to update a  table of a vendor
		sub.HandleFunc("PUT vendor/{id}/table/{table_id}", app.AuthMiddleware(http.HandlerFunc(app.requireVendorPermission(http.HandlerFunc(app.UpdateTableHandler)))))
		//to update a free a table of vendors
		sub.HandleFunc("PUT vendor/{id}/freetable/{table_id}", app.AuthMiddleware(http.HandlerFunc(app.requireVendorPermission(http.HandlerFunc(app.FreeCustomerTableHandler)))))
		//to delte a  table of a vendor
		sub.HandleFunc("DELETE vendor/{id}/tables/{table_id}", app.AuthMiddleware(http.HandlerFunc(app.requireVendorPermission(http.HandlerFunc(app.DeleteTableHandler)))))
		//to assign a table to a user by the users only
		sub.HandleFunc("PUT vendor/{id}/tables/{table_id}/needs-service", app.AuthMiddleware(http.HandlerFunc(app.UpdateTableNeedsServiceHandler)))
		//to free a table by the user who assigned it
		sub.HandleFunc("PUT vendor/{id}/tables/{table_id}/needs-serviceDone", app.AuthMiddleware(http.HandlerFunc(app.UpdateTableNeedsServiceHandler)))
		//to free a table by the user who assigned it
		sub.HandleFunc("PUT vendor/{id}/tables/{table_id}/freetable", app.AuthMiddleware(http.HandlerFunc(app.FreeTableHandler)))
		// Vendor routes
		sub.HandleFunc("GET vendors", app.AuthMiddleware(http.HandlerFunc(app.IndexVendorHandler)))
		sub.HandleFunc("GET vendors/{id}", app.AuthMiddleware(http.HandlerFunc(app.ShowVendorHandler)))
		sub.HandleFunc("POST vendors", app.AuthMiddleware(http.HandlerFunc(app.requireAdmin(http.HandlerFunc(app.CreateVendor)))))
		sub.HandleFunc("PUT vendors/{id}", app.AuthMiddleware(http.HandlerFunc(app.requireVendorPermission(http.HandlerFunc(app.UpdateVendorHandler)))))
		sub.HandleFunc("DELETE vendors/{id}", app.AuthMiddleware(http.HandlerFunc(app.requireAdmin(http.HandlerFunc(app.DeleteVendorHandler)))))
		sub.HandleFunc("GET vendortables/{id}", app.AuthMiddleware(http.HandlerFunc(app.GetVendorTablesHandler)))
		// Vendor Admin routes
		sub.HandleFunc("GET vendors/{id}/admins", app.AuthMiddleware(http.HandlerFunc(app.requireVendorPermission(http.HandlerFunc(app.GetVendorAdminsHandler)))))
		sub.HandleFunc("POST vendors/{id}/admins", app.AuthMiddleware(http.HandlerFunc(app.requireVendorPermission(http.HandlerFunc(app.CreateVendorAdminHandler)))))
		sub.HandleFunc("GET vendors/{id}/admins/{adminId}", app.AuthMiddleware(http.HandlerFunc(app.requireVendorPermission(http.HandlerFunc(app.GetVendorAdminHandler)))))
		sub.HandleFunc("PUT vendors/{id}/admins/{adminId}", app.AuthMiddleware(http.HandlerFunc(app.requireVendorPermission(http.HandlerFunc(app.UpdateVendorAdminHandler)))))
		sub.HandleFunc("DELETE vendors/{id}/admins/{adminId}", app.AuthMiddleware(http.HandlerFunc(app.requireVendorPermission(http.HandlerFunc(app.DeleteVendorAdminHandler)))))
		sub.HandleFunc("GET uservendors/{id}", app.AuthMiddleware(http.HandlerFunc(app.AuthorizeUserUpdate(http.HandlerFunc(app.GetUserVendor)))))
		//change the user's role
		sub.HandleFunc("PUT grantrole/{id}", app.AuthMiddleware(http.HandlerFunc(app.requireAdmin(http.HandlerFunc(app.GrantRole)))))
		//delete the user role
		sub.HandleFunc("DELETE revokerole", app.AuthMiddleware(http.HandlerFunc(app.requireAdmin(http.HandlerFunc(app.RevokeRoleHandler)))))
		sub.HandleFunc("GET userroles", app.AuthMiddleware(http.HandlerFunc(app.requireAdmin(http.HandlerFunc(app.IndexUserRoles)))))
		sub.HandleFunc("GET userroles/{id}", app.AuthMiddleware(http.HandlerFunc(app.requireAdmin(http.HandlerFunc(app.ShowUserRoleHandler)))))
		// Auth middleware applied per route
		sub.HandleFunc("GET me", app.AuthMiddleware(http.HandlerFunc(app.MeHandler)))
		sub.HandleFunc("GET users/{id}/vendors", app.AuthMiddleware(http.HandlerFunc(app.GetUserVendor)))
		sub.HandleFunc("POST orders", app.AuthMiddleware(http.HandlerFunc(app.CreateOrderHandler)))
		sub.HandleFunc("DELETE orders/{id}", app.AuthMiddleware(http.HandlerFunc(app.DeleteOrderHandler)))
		sub.HandleFunc("PUT orderscompleted/{id}", app.AuthMiddleware(http.HandlerFunc(app.UpdateOrderStatusHandler)))
		sub.HandleFunc("GET orders", app.AuthMiddleware(app.AuthorizeUserUpdate(http.HandlerFunc(app.GetOrdersHandler))))
		sub.HandleFunc("GET vendororders/{id}", app.AuthMiddleware(app.AuthorizeUserUpdate(http.HandlerFunc(app.GetVendorOrdersHandler))))
		sub.HandleFunc("POST orderitems", app.AuthMiddleware(http.HandlerFunc(app.CreateOrderItemHandler)))
		sub.HandleFunc("DELETE orderitems/{id}", app.AuthMiddleware(http.HandlerFunc(app.DeleteOrderItemHandler)))
		// add an item for a vendor
		sub.HandleFunc("POST vendor/{id}/items", app.AuthMiddleware(app.requireVendorPermission(http.HandlerFunc(app.CreateItemHandler))))
		// delete an item for a vendor
		sub.HandleFunc("DELETE vendor/{id}/items/{itemid}", app.AuthMiddleware(app.requireVendorPermission(http.HandlerFunc(app.DeleteItemHandler))))
		// get  items of a vendor
		sub.HandleFunc("GET vendor/{id}/items/{itemid}", app.AuthMiddleware(http.HandlerFunc(app.GetItemHandler)))
		sub.HandleFunc("GET vendor/{id}/items", app.AuthMiddleware(http.HandlerFunc(app.GetAllItemsHandler)))
		// update  items of a vendor
		sub.HandleFunc("GET vendor/{id}/itemscount", app.AuthMiddleware(http.HandlerFunc(app.GetAllItemsCountHandler)))
		sub.HandleFunc("PUT vendor/{id}/items/{itemid}", app.AuthMiddleware(app.requireVendorPermission(http.HandlerFunc(app.UpdateItemHandler))))
		sub.HandleFunc("POST cartitems", app.AuthMiddleware(http.HandlerFunc(app.CreateCartItemHandler)))
		sub.HandleFunc("GET cartitems", app.AuthMiddleware(http.HandlerFunc(app.GetCartItemswithimage)))
		sub.HandleFunc("DELETE cartitems/{id}", app.AuthMiddleware(http.HandlerFunc(app.DeleteCartItemHandler)))
		sub.HandleFunc("PUT cartitems/{id}", app.AuthMiddleware((http.HandlerFunc(app.UpdateCartItemHandler))))
		sub.HandleFunc("POST carts", app.AuthMiddleware(http.HandlerFunc(app.CreateCartHandler)))
		sub.HandleFunc("DELETE carts/{id}", app.AuthMiddleware(app.requireAdmin(http.HandlerFunc(app.DeleteCartHandler))))
		sub.HandleFunc("PUT carts/{id}", app.AuthMiddleware(app.AuthorizeUserUpdate(http.HandlerFunc(app.UpdateCartHandler))))
		sub.HandleFunc("GET carts", app.AuthMiddleware(app.AuthorizeUserUpdate(http.HandlerFunc(app.GetCartHandler))))
		sub.HandleFunc("POST checkout", app.AuthMiddleware(app.AuthorizeUserUpdate(http.HandlerFunc(app.CheckoutHandler))))
	})

	return r
}
