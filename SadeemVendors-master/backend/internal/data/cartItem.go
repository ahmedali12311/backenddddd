package data

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// CartItem represents an item in the cart.
type CartItem struct {
	CartID   uuid.UUID `db:"cart_id" json:"cart_id"`
	ItemID   uuid.UUID `db:"item_id" json:"item_id"`
	Quantity int       `db:"quantity" json:"quantity"`
}
type CartItemWithNameAndImg struct {
	CartItem
	Name string  `json:"name"`
	Img  *string `json:"img"`
}
type CartItemDB struct {
	db *sqlx.DB
}

func (c *CartItemDB) InsertCartItem(cartItem *CartItem) error {
	// Check if the item already exists in the cart
	exists, err := c.ItemExistsInCart(cartItem.CartID, cartItem.ItemID)
	if err != nil {
		return fmt.Errorf("error checking if item exists: %v", err)
	}
	if exists {
		return ErrItemAlreadyInserted
	}

	query, args, err := QB.Insert("cart_items").
		Columns(strings.Join(cartItemsColumns, ",")).
		Values(cartItem.CartID, cartItem.ItemID, cartItem.Quantity).
		ToSql()
	if err != nil {
		return fmt.Errorf("error building insert query: %v", err)
	}

	_, err = c.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("error while inserting cart item: %v", err)
	}
	return nil
}

// UpdateCartItem updates the quantity of an item in the cart.
func (c *CartItemDB) Updatecartitem(cartItem *CartItem) error {
	query, args, err := QB.Update("cart_items").
		Set("quantity", cartItem.Quantity).
		Where(squirrel.Eq{"cart_id": cartItem.CartID, "item_id": cartItem.ItemID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("error building update query: %v", err)
	}

	_, err = c.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("error while updating cart item: %v", err)
	}

	return nil
}

func (c *CartItemDB) DeleteCartItem(ctx context.Context, cartID, itemID uuid.UUID) error {
	var cartItem CartItem

	// Step 1: Delete the item and return the deleted item
	query, args, err := QB.Delete("cart_items").
		Where(squirrel.Eq{"cart_id": cartID, "item_id": itemID}).
		Suffix(fmt.Sprintf("RETURNING %s", strings.Join([]string{"cart_id", "item_id", "quantity"}, ", "))).
		ToSql()
	if err != nil {
		return fmt.Errorf("error building delete query: %v", err)
	}

	err = c.db.QueryRowxContext(ctx, query, args...).StructScan(&cartItem)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil // No rows affected means the item was not found and thus already deleted
		}
		return fmt.Errorf("error deleting cart item: %v", err)
	}

	return nil
}

// GetCartItems retrieves all items in a specified cart.
func (c *CartItemDB) GetCartItems(cartID uuid.UUID) ([]CartItem, error) {
	var items []CartItem

	query, args, err := QB.Select(cartItemsColumns...).
		From("cart_items").
		Where(squirrel.Eq{"cart_id": cartID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building query: %v", err)
	}

	// Use sqlx's Select method to scan results directly into items slice
	err = c.db.Select(&items, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no items found for cart ID: %s", cartID)
		}
		return nil, fmt.Errorf("error while querying cart items: %v", err)
	}

	return items, nil
}
func (m *CartItemDB) DeleteCartItems(cartID uuid.UUID) error {
	query, args, err := squirrel.Delete("cart_items").Where(squirrel.Eq{"cart_id": cartID}).ToSql()
	if err != nil {
		return err
	}
	_, err = m.db.Exec(query, args...)
	return err
}
func (c *CartItemDB) GetCartItemswithimage(cartID uuid.UUID) ([]CartItemWithNameAndImg, error) {
	var items []CartItemWithNameAndImg

	query, args, err := QB.Select(
		"cart_items.cart_id",
		"cart_items.item_id",
		"cart_items.quantity",
		"items.name",
		fmt.Sprintf("CASE WHEN NULLIF(img, '') IS NOT NULL THEN FORMAT('%s/%%s', img) ELSE NULL END AS img", Domain)).
		From("cart_items").
		Join("items ON cart_items.item_id = items.id").
		Where(squirrel.Eq{"cart_id": cartID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building query: %v", err)
	}

	err = c.db.Select(&items, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no items found for cart ID: %s", cartID)
		}
		return nil, fmt.Errorf("error while querying cart items: %v", err)
	}

	return items, nil
}

func (c *CartItemDB) ItemExistsInCart(cartID, itemID uuid.UUID) (bool, error) {
	var count int

	query, args, err := QB.Select("COUNT(*)").
		From("cart_items").
		Where(squirrel.Eq{"cart_id": cartID, "item_id": itemID}).
		ToSql()
	if err != nil {
		return false, fmt.Errorf("error building query: %v", err)
	}

	err = c.db.Get(&count, query, args...)
	if err != nil {
		return false, fmt.Errorf("error checking item existence: %v", err)
	}

	return count > 0, nil
}
