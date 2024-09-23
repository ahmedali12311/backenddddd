package data

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Cart represents a shopping cart.
type Cart struct {
	ID         uuid.UUID `db:"id" json:"id"`
	TotalPrice float64   `db:"total_price" json:"total_price"`
	Quantity   int       `db:"quantity" json:"quantity"`
	VendorID   uuid.UUID `db:"vendor_id" json:"vendor_id"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UpdatedAt  time.Time `db:"updated_at" json:"updated_at"`
}

type CartDB struct {
	db *sqlx.DB
}

func (db *CartDB) GetCart(userID uuid.UUID) (*Cart, error) {
	var cart Cart

	// Build the query using squirrel
	query, args, err := QB.Select(cartsColumns...).
		From("carts").
		Where(squirrel.Eq{"id": userID}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("error building query: %v", err)
	}

	// Execute the query
	err = db.db.QueryRowx(query, args...).StructScan(&cart)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrRecordNotFound // Ensure this error is defined in your data package
		}
		return nil, fmt.Errorf("error while querying cart: %v", err)
	}

	return &cart, nil
}
func (c *CartDB) InsertCart(cart *Cart) error {
	query, args, err := QB.Insert("carts").
		Columns("id", "vendor_id").
		Values(cart.ID, cart.VendorID).
		ToSql()
	if err != nil {
		return err
	}
	_, err = c.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("error while inserting cart: %v", err)
	}
	return nil
}

func (c *CartDB) DeleteCart(cartID uuid.UUID) error {
	query, args, err := QB.Delete("carts").
		Where(squirrel.Eq{"id": cartID}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = c.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("error while deleting cart: %v", err)
	}
	return nil
}

func (db *CartDB) UpdateCart(cart *Cart) error {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	updateQuery := psql.Update("carts").
		Set("total_price", cart.TotalPrice).
		Set("quantity", cart.Quantity).
		Set("vendor_id", cart.VendorID).
		Set("updated_at", time.Now()).
		Where(squirrel.Eq{"id": cart.ID})

	sql, args, err := updateQuery.ToSql()
	if err != nil {
		return err
	}

	_, err = db.db.Exec(sql, args...)
	if err != nil {
		return err
	}

	return nil
}
