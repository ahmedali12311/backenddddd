package data

import (
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Transaction struct {
	tx *sqlx.Tx
}

func (m *Model) BeginTransaction() (*Transaction, error) {
	tx, err := m.ItemDB.db.Beginx()
	if err != nil {
		return nil, err
	}
	return &Transaction{tx: tx}, nil
}

func (t *Transaction) Commit() error {
	return t.tx.Commit()
}

func (t *Transaction) Rollback() error {
	return t.tx.Rollback()
}

func (t *Transaction) InsertOrder(order *Order) error {
	query, args, err := QB.Insert("orders").
		Columns(ordersColumns...).
		Values(order.ID, order.TotalOrderCost, order.CustomerID, order.VendorID, order.Status, order.CreatedAt, order.UpdatedAt).
		ToSql()
	if err != nil {
		return err
	}

	_, err = t.tx.Exec(query, args...)
	return err
}

// InsertOrderItem inserts a new order item into the database.
func (t *Transaction) InsertOrderItem(orderItem *OrderItem) error {
	query, args, err := QB.Insert("order_items").
		Columns(orderItemsColumns...).
		Values(orderItem.ID, orderItem.OrderID, orderItem.ItemID, orderItem.Quantity, orderItem.Price).
		ToSql()
	if err != nil {
		return fmt.Errorf("error building insert order item query: %v", err)
	}

	_, err = t.tx.Exec(query, args...)
	return err
}

// DeleteCart deletes a cart from the database.
func (t *Transaction) DeleteCart(cartID uuid.UUID) error {
	query, args, err := QB.Delete("carts").
		Where(squirrel.Eq{"id": cartID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("error building delete cart query: %v", err)
	}

	_, err = t.tx.Exec(query, args...)
	return err
}

// DeleteCartItems deletes all items from a cart.
func (t *Transaction) DeleteCartItems(cartID uuid.UUID) error {
	query, args, err := QB.Delete("cart_items").
		Where(squirrel.Eq{"cart_id": cartID}).
		ToSql()
	if err != nil {
		return fmt.Errorf("error building delete cart items query: %v", err)
	}

	_, err = t.tx.Exec(query, args...)
	return err
}
