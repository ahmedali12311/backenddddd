package data

import (
	"errors"
	"fmt"
	"os"

	"github.com/Masterminds/squirrel"
	"github.com/jmoiron/sqlx"
	_ "github.com/joho/godotenv/autoload"
)

var (
	ErrRecordNotFound        = errors.New("record not found")
	ErrDuplicatedKey         = errors.New("user already have the value")
	ErrDuplicatedRole        = errors.New("user Already have the role")
	ErrHasRole               = errors.New("user Already has a role")
	ErrHasNoRoles            = errors.New("user Has no roles")
	ErrForeignKeyViolation   = errors.New("foreign key constraint violation")
	ErrUserNotFound          = errors.New("user Not Found")
	ErrUserAlreadyhaveatable = errors.New("user already have a table")
	ErrUserHasNoTable        = errors.New("user has no table")
	ErrItemAlreadyInserted   = errors.New("item already inserted! ")
	ErrInvalidQuantity       = errors.New("requested quantity is not available")

	QB     = squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	Domain = os.Getenv("DOMAIN")

	user_columns = []string{
		"id",
		"name",
		"email",
		"password",
		"phone",
		"created_at",
		"updated_at",
		fmt.Sprintf("CASE WHEN NULLIF(img, '') IS NOT NULL THEN FORMAT('%s/%%s', img) ELSE NULL END AS img", Domain),
	}
	vendors_columns = []string{
		"id",
		"name",
		"description",
		"subscription_end",
		"subscription_days",
		"is_visible",
		"created_at",
		"updated_at",
		fmt.Sprintf("CASE WHEN NULLIF(img, '') IS NOT NULL THEN FORMAT('%s/%%s', img) ELSE NULL END AS img", Domain),
	}
	user_roles = []string{
		"user_id",
		"role_id",
	}
	tableColumns     = []string{"id", "name", "vendor_id", "customer_id", "is_available", "is_needs_service"}
	cartItemsColumns = []string{
		"cart_id", "item_id", "quantity",
	}

	cartsColumns = []string{
		"id", "total_price", "quantity", "vendor_id", "created_at", "updated_at",
	}

	orderItemsColumns = []string{
		"id", "order_id", "item_id", "quantity", "price",
	}

	ordersColumns = []string{
		"id", "total_order_cost", "customer_id", "vendor_id", "status", "created_at", "updated_at",
	}

	itemsColumns = []string{
		"id",
		"vendor_id",
		"name",
		"price",
		"quantity",
		"discount",
		"discount_expiry",
		"created_at",
		"updated_at",
		fmt.Sprintf("CASE WHEN NULLIF(img, '') IS NOT NULL THEN FORMAT('%s/%%s', img) ELSE NULL END AS img", Domain),
	}
)

type Model struct {
	UserDB        UserDB
	TableDB       TableDB
	VendorDB      VendorDB
	UserRoleDB    UserRoleDB
	VendorAdminDB VendorAdminDB
	CartItemDB    CartItemDB
	CartDB        CartDB
	OrderItemDB   OrderItemDB
	OrderDB       OrderDB
	ItemDB        ItemDB
	TransactionDB Transaction
}

func NewModels(db *sqlx.DB) Model {
	tx, err := db.Beginx()
	if err != nil {
		return Model{}
	}
	return Model{
		UserDB:        UserDB{db},
		TableDB:       TableDB{db},
		VendorDB:      VendorDB{db},
		UserRoleDB:    UserRoleDB{db},
		VendorAdminDB: VendorAdminDB{db},
		CartItemDB:    CartItemDB{db},
		CartDB:        CartDB{db},
		OrderItemDB:   OrderItemDB{db},
		OrderDB:       OrderDB{db},
		ItemDB:        ItemDB{db},
		TransactionDB: Transaction{tx},
	}
}
