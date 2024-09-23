package data

import (
	"fmt"
	"project/utils/validator"
	"strconv"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type OrderDetails struct {
	ID             uuid.UUID `json:"id"`
	TotalOrderCost float64   `json:"total_order_cost"`
	VendorName     string    `json:"vendor_name"`
	VendorID       uuid.UUID `json:"-"`
	UserName       string    `json:"user_name"`
	ItemNames      []string  `json:"item_names"`
	ItemPrices     []float64 `json:"item_prices"`
	ItemQuantities []int     `json:"item_quantities"` // New field for item quantities
	Status         string    `json:"status"`
	TableID        uuid.UUID `json:"table_id"`
	TableName      string    `json:"table_name"`
}

// Order represents an order.
type Order struct {
	ID             uuid.UUID `db:"id" json:"id"`
	TotalOrderCost float64   `db:"total_order_cost" json:"total_order_cost"`
	CustomerID     uuid.UUID `db:"customer_id" json:"customer_id"`
	VendorID       uuid.UUID `db:"vendor_id" json:"vendor_id"`
	Status         string    `db:"status" json:"status"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

type OrderDB struct {
	db *sqlx.DB
}

func ValidatingOrder(v *validator.Validator, order *Order, fields ...string) {
	for _, field := range fields {
		switch field {
		case "customer_id":
			v.Check(order.CustomerID != uuid.Nil, "customer_id", "Customer ID is required")
		case "vendor_id":
			v.Check(order.VendorID != uuid.Nil, "vendor_id", "Vendor ID is required")
		case "status":
			v.Check(order.Status != "", "status", "Order status is required")
			v.Check(order.Status == "completed " || order.Status == "preparing", "status", "Order must be of 2 status")

		case "total_order_cost":
			v.Check(order.TotalOrderCost >= 0, "total_order_cost", "Total order cost must be a non-negative number")
		}
	}
}
func (o *OrderDB) GetOrders(customerID uuid.UUID, tableID uuid.UUID) ([]OrderDetails, error) {
	query, args, err := QB.Select(
		"o.id",
		"o.total_order_cost",
		"v.name AS vendor_name",
		"v.id AS vendor_id",
		"c.name AS user_name",
		"array_agg(i.name) AS item_names",
		"array_agg(i.price::text) AS item_prices",
		"array_agg(oi.quantity) AS item_quantities", // Aggregate item quantities
		"o.status",
		"t.id AS table_id",
		"t.name AS table_name",
	).
		From("orders o").
		Join("vendors v ON o.vendor_id = v.id").
		Join("users c ON o.customer_id = c.id").
		Join("order_items oi ON o.id = oi.order_id").
		Join("items i ON oi.item_id = i.id").
		Join("tables t ON o.customer_id = t.customer_id").
		Where(squirrel.Eq{"o.customer_id": customerID, "t.id": tableID}).
		GroupBy("o.id, o.total_order_cost, v.name, v.id, c.name, o.status, t.id, t.name").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := o.db.Queryx(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []OrderDetails
	for rows.Next() {
		var order OrderDetails
		var itemPricesStr []string
		var itemQuantitiesStr []string

		err := rows.Scan(
			&order.ID,
			&order.TotalOrderCost,
			&order.VendorName,
			&order.VendorID,
			&order.UserName,
			pq.Array(&order.ItemNames),
			pq.Array(&itemPricesStr),
			pq.Array(&itemQuantitiesStr), // Scan item quantities
			&order.Status,
			&order.TableID,
			&order.TableName,
		)
		if err != nil {
			return nil, err
		}

		// Convert item prices from strings to floats
		var itemPrices []float64
		for _, priceStr := range itemPricesStr {
			price, err := strconv.ParseFloat(priceStr, 64)
			if err != nil {
				return nil, err
			}
			itemPrices = append(itemPrices, price)
		}
		order.ItemPrices = itemPrices

		// Convert item quantities from strings to ints
		var itemQuantities []int
		for _, quantityStr := range itemQuantitiesStr {
			quantity, err := strconv.Atoi(quantityStr)
			if err != nil {
				return nil, err
			}
			itemQuantities = append(itemQuantities, quantity)
		}
		order.ItemQuantities = itemQuantities

		orders = append(orders, order)
	}
	return orders, nil
}

func (o *OrderDB) InsertOrder(order *Order) error {
	query, args, err := QB.Insert("orders").
		Columns(strings.Join(ordersColumns, ",")).
		Values(order.ID, order.TotalOrderCost, order.CustomerID, order.VendorID, order.Status).
		ToSql()
	if err != nil {
		return err
	}
	_, err = o.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("error while inserting order: %v", err)
	}
	return nil
}

func (o *OrderDB) DeleteOrder(orderID uuid.UUID) error {
	query, args, err := QB.Delete("orders").
		Where(squirrel.Eq{"id": orderID}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = o.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("error while deleting order: %v", err)
	}
	return nil
}
func (o *OrderDB) GetVendorOrders(vendorID uuid.UUID) ([]Order, error) {
	query, args, err := QB.Select(strings.Join(ordersColumns, ",")).
		From("orders").
		Where(squirrel.Eq{"vendor_id": vendorID}).
		OrderBy("created_at ASC").
		ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := o.db.Queryx(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		err := rows.StructScan(&order)
		if err != nil {
			return nil, err
		}
		orders = append(orders, order)
	}
	return orders, nil
}

// UpdateOrder updates the order status to "completed".
func (o *OrderDB) UpdateOrder(orderID uuid.UUID, status string) error {
	if status != "completed" {
		return fmt.Errorf("only 'completed' status is allowed")
	}

	query, args, err := QB.Update("orders").
		Set("status", status).
		Where(squirrel.Eq{"id": orderID}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = o.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("error while updating order: %v", err)
	}
	return nil
}

// DeleteCompletedOrder deletes the order if its status is "completed".
func (o *OrderDB) DeleteCompletedOrder(orderID uuid.UUID) error {
	// Get the order by ID
	order, err := o.GetOrder(orderID)
	if err != nil {
		return fmt.Errorf("error while getting order: %v", err)
	}

	// Check if the order is completed
	if order.Status != "completed" {
		return fmt.Errorf("order is not completed")
	}

	// Delete the order
	query, args, err := QB.Delete("orders").
		Where(squirrel.Eq{"id": orderID}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = o.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("error while deleting order: %v", err)
	}
	return nil
}

func (o *OrderDB) GetOrder(orderID uuid.UUID) (*Order, error) {
	query, args, err := QB.Select(strings.Join(ordersColumns, ",")).
		From("orders").
		Where(squirrel.Eq{"id": orderID}).
		ToSql()
	if err != nil {
		return nil, err
	}
	row, err := o.db.Queryx(query, args...)
	if err != nil {
		return nil, err
	}
	defer row.Close()

	if !row.Next() {
		return nil, fmt.Errorf("order not found")
	}

	var order Order
	err = row.StructScan(&order)
	if err != nil {
		return nil, err
	}
	return &order, nil
}
func (o *OrderDB) DeleteUserOrders(customerID uuid.UUID) error {
	query, args, err := QB.Delete("orders").
		Where(squirrel.Eq{"customer_id": customerID}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = o.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("error while deleting user orders: %v", err)
	}
	return nil
}
