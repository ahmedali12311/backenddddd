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

// Table represents a table in the database.
type Table struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	Name            string     `db:"name" json:"name"`
	VendorID        uuid.UUID  `db:"vendor_id" json:"vendor_id"`
	CustomerID      *uuid.UUID `db:"customer_id,omitempty" json:"customer_id,omitempty"`
	IsAvailable     bool       `db:"is_available" json:"is_available"`
	IsNeedsServices bool       `db:"is_needs_service" json:"is_needs_service"`
}

// TableDB wraps a sqlx.DB connection pool.
type TableDB struct {
	DB *sqlx.DB
}

// GetTables retrieves all tables from the database.
func (db *TableDB) GetTables(ctx context.Context) (*[]Table, error) {
	var tables []Table
	query, args, err := QB.Select(strings.Join(tableColumns, ",")).From("tables").ToSql()
	if err != nil {
		return nil, err
	}
	err = db.DB.SelectContext(ctx, &tables, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &tables, nil
}

// GetTable retrieves a table by its ID.
func (db *TableDB) GetTable(ctx context.Context, id uuid.UUID) (*Table, error) {
	var table Table
	query, args, err := QB.Select(strings.Join(tableColumns, ",")).From("tables").Where(squirrel.Eq{"id": id}).ToSql()
	if err != nil {
		return nil, err
	}
	err = db.DB.QueryRowxContext(ctx, query, args...).StructScan(&table)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &table, nil
}

// Insert inserts a new table into the database.
func (db *TableDB) Insert(ctx context.Context, table *Table) error {
	query, args, err := QB.
		Insert("tables").
		Columns("name", "vendor_id", "customer_id", "is_available", "is_needs_service").
		Values(table.Name, table.VendorID, table.CustomerID, table.IsAvailable, table.IsNeedsServices).
		Suffix(fmt.Sprintf("RETURNING %s", strings.Join(tableColumns, ", "))).
		ToSql()
	if err != nil {
		return err
	}

	err = db.DB.QueryRowxContext(ctx, query, args...).StructScan(table)
	if err != nil {
		return err
	}
	return nil
}

// DeleteTable deletes a table by its ID.
func (db *TableDB) DeleteTable(ctx context.Context, id uuid.UUID) (*Table, error) {
	var table Table
	query, args, err := QB.Delete("tables").Where(squirrel.Eq{"id": id}).Suffix(fmt.Sprintf("RETURNING %s", strings.Join(tableColumns, ", "))).ToSql()
	if err != nil {
		return nil, err
	}
	err = db.DB.QueryRowxContext(ctx, query, args...).StructScan(&table)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &table, nil
}

// Update updates an existing table in the database.
func (db *TableDB) Update(ctx context.Context, table *Table) error {
	query, args, err := QB.Update("tables").
		Set("name", table.Name).
		Set("is_available", table.IsAvailable).
		Set("is_needs_service", table.IsNeedsServices).
		Where(squirrel.Eq{"id": table.ID}).
		Suffix(fmt.Sprintf("RETURNING %s", strings.Join(tableColumns, ", "))).
		ToSql()
	if err != nil {
		return err
	}

	result, err := db.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("table not found")
	}
	return nil
}

func (db *TableDB) CountTablesForVendor(ctx context.Context, vendorID uuid.UUID) (int, error) {
	var count int

	query, args, err := QB.Select("COUNT(*)").
		From("tables").
		Where(squirrel.Eq{"vendor_id": vendorID}).ToSql()
	if err != nil {
		return 0, err
	}

	err = db.DB.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		fmt.Printf("Error executing query: %v\nQuery: %s\nArgument: %v\n", err, query, vendorID)
		return 0, err
	}

	return count, nil
}
func (v *TableDB) GetVendorTables(ctx context.Context, vendorID uuid.UUID) ([]Table, error) {
	var tables []Table

	// Initialize squirrel query builder
	query, args, err := QB.Select(tableColumns...).From("tables").
		Where(squirrel.Eq{"vendor_id": vendorID}).
		ToSql()
	if err != nil {
		return nil, err
	}

	// Execute the query and scan the results into the tables slice
	err = v.DB.SelectContext(ctx, &tables, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecordNotFound
		}
		fmt.Printf("Error executing query: %v\nQuery: %s\nArgs: %v\n", err, query, args)
		return nil, err
	}

	return tables, nil
}

// Assign a customer to a table
func (db *TableDB) AssignCustomer(ctx context.Context, tableID, customerID uuid.UUID) error {
	err := db.DoCustomerHavetable(ctx, customerID)
	if err != nil {
		if err == ErrUserAlreadyhaveatable {
			return ErrUserAlreadyhaveatable
		}
	}
	// Update table to assign customer and set is_available to false
	query, args, err := QB.Update("tables").
		Set("customer_id", customerID).
		Set("is_available", false).
		Where(squirrel.Eq{"id": tableID}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = db.DB.ExecContext(ctx, query, args...)
	if err != nil {
		return err

	}
	return nil

}

// Free a table
func (db *TableDB) FreeTable(ctx context.Context, tableID uuid.UUID) error {
	// Update table to remove customer and set is_available to true
	query, args, err := QB.Update("tables").
		Set("customer_id", nil).
		Set("is_available", true).
		Where(squirrel.Eq{"id": tableID}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = db.DB.ExecContext(ctx, query, args...)

	return err
}

func (db *TableDB) FreeTableVendor(ctx context.Context, tableID uuid.UUID, vendorID uuid.UUID) error {

	// Update table to remove customer and set is_available to true
	query, args, err := QB.Update("tables").
		Set("customer_id", nil).
		Set("is_available", true).
		Where(squirrel.Eq{"id": tableID, "vendor_id": vendorID}).
		ToSql()
	if err != nil {
		return err
	}

	_, err = db.DB.ExecContext(ctx, query, args...)

	return err
}

// GetTable retrieves a table by its ID.
func (db *TableDB) DoCustomerHavetable(ctx context.Context, customerid uuid.UUID) error {
	var table Table
	query, args, err := QB.Select(strings.Join(tableColumns, ",")).From("tables").Where(squirrel.Eq{"customer_id": customerid}).ToSql()
	if err != nil {
		return err
	}
	err = db.DB.QueryRowxContext(ctx, query, args...).StructScan(&table)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil
		}
		return err
	}
	return ErrUserAlreadyhaveatable
}

// GetTable retrieves a table by its ID.
func (db *TableDB) GetCustomertable(ctx context.Context, customerid uuid.UUID) (*Table, error) {
	var table Table
	query, args, err := QB.Select(strings.Join(tableColumns, ",")).From("tables").Where(squirrel.Eq{"customer_id": customerid}).ToSql()
	if err != nil {
		return nil, err
	}
	err = db.DB.QueryRowxContext(ctx, query, args...).StructScan(&table)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserHasNoTable
		}
		return nil, err
	}
	return &table, nil
}
