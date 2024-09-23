package data

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"project/utils"
	"project/utils/validator"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Item represents an item for sale.
type Item struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	VendorID       uuid.UUID  `db:"vendor_id" json:"vendor_id"`
	Name           string     `db:"name" json:"name"`
	Price          float64    `db:"price" json:"price"`
	Discount       float64    `db:"discount" json:"discount"`
	DiscountExpiry *time.Time `db:"discount_expiry" json:"discount_expiry"`
	Quantity       int        `db:"quantity" json:"quantity"`
	Img            *string    `db:"img" json:"img"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updated_at"`
}
type ItemDB struct {
	db *sqlx.DB
}

// ValidatingItem performs validations on the Item struct based on provided fields.
func ValidatingItem(v *validator.Validator, item *Item, fields ...string) {
	for _, field := range fields {
		switch field {
		case "quantity":
			if item.Quantity < 0 {
				v.Check(item.Quantity >= 0, "quantity", "Quantity must be non-negative")
			}
		case "name":
			if item.Name != "" {
				v.Check(len(item.Name) <= 20, "name", "Name must be less than 50 characters")
			}
		case "price":
			if item.Price <= 0 {
				v.Check(item.Price > 0, "price", "Price must be greater than zero")
			}
		case "discount":
			if item.Discount != 0 {
				v.Check(item.Discount > 0, "discount", "Discount must not be negative")
				v.Check(item.Discount < item.Price, "discount", "Discount must be less than the price!")
			}
		case "discount_expiry":
			if item.Discount != 0 && item.DiscountExpiry == nil {
				v.Check(item.DiscountExpiry != nil, "discount_expiry", "Discount expiry date is required when a discount is provided correctly!")
			}
		}
	}
}

func (i *ItemDB) InsertItem(item *Item) error {
	query, args, err := QB.Insert("items").
		Columns("vendor_id", "name", "price", "img", "discount", "discount_expiry", "quantity").
		Values(item.VendorID, item.Name, item.Price, item.Img, item.Discount, item.DiscountExpiry, item.Quantity).
		Suffix("RETURNING " + strings.Join(itemsColumns, ", ")).
		ToSql()
	if err != nil {
		return err
	}
	err = i.db.QueryRowx(query, args...).StructScan(item)
	if err != nil {
		return fmt.Errorf("error while inserting item: %v", err)
	}
	return nil
}

func (i *ItemDB) DeleteItem(itemID uuid.UUID) error {
	item, err := i.GetItem(itemID)
	if err != nil {
		if err == ErrRecordNotFound {
			return ErrRecordNotFound
		}
	}
	query, args, err := QB.Delete("items").
		Where(squirrel.Eq{"id": itemID}).
		Suffix(fmt.Sprintf("RETURNING %s", strings.Join(itemsColumns, ", "))).
		ToSql()
	if err != nil {
		return err
	}
	err = i.db.QueryRowx(query, args...).StructScan(item)
	if err != nil {
		if err == sql.ErrNoRows {
			return ErrRecordNotFound
		}
		return err

	}
	if item.Img != nil {
		imgfile := strings.TrimPrefix(*item.Img, Domain+"/")
		// Check if the file exists
		if _, err := os.Stat(imgfile); err == nil {
			// File exists, attempt to delete it
			err = utils.DeleteImageFile(imgfile)
			if err != nil {
				return fmt.Errorf("failed to delete file %s: %w", imgfile, err)
			}
		} else if os.IsNotExist(err) {
			// File does not exist, log a warning but do not treat it as a fatal error
			fmt.Printf("Warning: image file %s does not exist\n", imgfile)
		} else {
			// Handle other potential errors from os.Stat
			return fmt.Errorf("failed to check file %s: %w", imgfile, err)
		}
	}
	return nil
}
func (i *ItemDB) GetAllItems(vendorID uuid.UUID, filters utils.Filters) ([]Item, error) {
	var items []Item
	queryBuilder := QB.Select(itemsColumns...).
		From("items").
		Where(squirrel.Eq{"vendor_id": vendorID})

	// Add search functionality if needed
	if filters.Search != "" {
		queryBuilder = queryBuilder.Where(squirrel.Like{"name": "%" + filters.Search + "%"})
	}

	// Sorting
	if filters.Sort != "" {
		queryBuilder = queryBuilder.OrderBy(filters.Sort)
	}

	offset := (filters.Page - 1) * filters.PageSize
	queryBuilder = queryBuilder.Limit(uint64(filters.PageSize)).Offset(uint64(offset))

	// Build SQL query
	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	err = i.db.Select(&items, query, args...)
	if err != nil {
		return nil, fmt.Errorf("error while retrieving items: %v", err)
	}
	return items, nil
}

func (i *ItemDB) GetItem(itemID uuid.UUID) (*Item, error) {
	var item Item
	query, args, err := QB.Select(itemsColumns...).
		From("items").
		Where(squirrel.Eq{"id": itemID}).
		ToSql()
	if err != nil {
		return nil, err
	}
	err = i.db.Get(&item, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}

	return &item, nil
}
func (i *ItemDB) UpdateItem(item *Item) error {
	query, args, err := QB.Update("items").
		SetMap(map[string]interface{}{
			"name":            item.Name,
			"price":           item.Price,
			"img":             item.Img,
			"discount":        item.Discount,
			"discount_expiry": item.DiscountExpiry,
			"quantity":        item.Quantity,
			"updated_at":      time.Now(),
		}).
		Where(squirrel.Eq{"id": item.ID}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = i.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("error while updating item: %v", err)
	}
	return nil
}
func (i *ItemDB) GetAllItemsCount(vendorID uuid.UUID) (int64, error) {
	var items int64
	queryBuilder := QB.Select("COUNT(*)").
		From("items").
		Where(squirrel.Eq{"vendor_id": vendorID})

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return 0, err
	}

	err = i.db.Get(&items, query, args...)
	if err != nil {
		return 0, fmt.Errorf("error while retrieving items: %v", err)
	}
	return items, nil
}
func (db *ItemDB) GetItemPrice(itemID uuid.UUID) (*Item, error) {
	var price Item

	// Build the query using squirrel
	query, args, err := QB.Select("price,discount").
		From("items").
		Where(squirrel.Eq{"id": itemID}).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building query: %v", err)
	}

	// Execute the query
	err = db.db.QueryRow(query, args...).Scan(&price.Price, &price.Discount)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("item not found: %s", itemID)
		}
		return nil, fmt.Errorf("error while querying item price: %v", err)
	}

	return &price, nil
}
func (i *ItemDB) GetVendorID(itemID uuid.UUID) (uuid.UUID, error) {
	item, err := i.GetItem(itemID) // Reuse the existing GetItem method
	if err != nil {
		return uuid.Nil, err
	}
	return item.VendorID, nil
}
func (i *ItemDB) IsStockAvailable(itemID uuid.UUID, quantity int) (bool, error) {
	item, err := i.GetItem(itemID)
	if err != nil {
		return false, err
	}
	return item.Quantity >= quantity, nil
}
