package data

import (
	"fmt"
	"strings"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// OrderItem represents an item in an order.
type OrderItem struct {
	ID       uuid.UUID `db:"id" json:"id"`
	OrderID  uuid.UUID `db:"order_id" json:"order_id"`
	ItemID   uuid.UUID `db:"item_id" json:"item_id"`
	Quantity int       `db:"quantity" json:"quantity"`
	Price    float64   `db:"price" json:"price"`
}

type OrderItemDB struct {
	db *sqlx.DB
}

func (o *OrderItemDB) InsertOrderItem(orderItem *OrderItem) error {
	query, args, err := QB.Insert("order_items").
		Columns(strings.Join(orderItemsColumns, ",")).
		Values(orderItem.ID, orderItem.OrderID, orderItem.ItemID, orderItem.Quantity, orderItem.Price).
		ToSql()
	if err != nil {
		return err
	}
	_, err = o.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("error while inserting order item: %v", err)
	}
	return nil
}

func (o *OrderItemDB) DeleteOrderItem(orderItemID uuid.UUID) error {
	query, args, err := QB.Delete("order_items").
		Where(squirrel.Eq{"id": orderItemID}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = o.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("error while deleting order item: %v", err)
	}
	return nil
}
