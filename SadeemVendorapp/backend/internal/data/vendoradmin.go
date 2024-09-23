package data

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// VendorAdmin represents a vendor admin record.
type VendorAdmin struct {
	UserID   uuid.UUID `db:"user_id" json:"user_id"`
	VendorID uuid.UUID `db:"vendor_id" json:"vendor_id"`
}
type VendorAdminUser struct {
	UserID   uuid.UUID `db:"user_id" json:"user_id"`
	VendorID uuid.UUID `db:"vendor_id" json:"vendor_id"`
	Email    string    `db:"email"      json:"email"`
}

// VendorAdminDB wraps a sqlx.DB connection pool for vendor admins.
type VendorAdminDB struct {
	db *sqlx.DB
}

// InsertVendorAdmin inserts a new vendor admin record into the database.
func (v *VendorAdminDB) InsertVendorAdmin(ctx context.Context, vendor VendorAdmin) (*VendorAdmin, error) {
	query, args, err := QB.Insert("vendor_admins").Columns("user_id", "vendor_id").
		Values(vendor.UserID, vendor.VendorID).
		Suffix("RETURNING user_id, vendor_id").ToSql()
	if err != nil {
		return nil, err
	}

	err = v.db.QueryRowxContext(ctx, query, args...).StructScan(&vendor)
	if err != nil {
		// Check for unique constraint violation (PostgreSQL error code for unique violation is 23505)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return nil, ErrDuplicatedRole
		}
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23503" {
			return nil, ErrForeignKeyViolation
		}
		return nil, err
	}
	return &vendor, nil
}

// GetVendorAdmin retrieves a vendor admin record by user_id and vendor_id.
func (v *VendorAdminDB) GetVendorAdmin(ctx context.Context, userID, vendorID uuid.UUID) (*VendorAdmin, error) {
	var vendorAdmin VendorAdmin
	query, args, err := QB.Select("user_id, vendor_id").
		From("vendor_admins").
		Where(squirrel.Eq{"user_id": userID, "vendor_id": vendorID}).
		ToSql()
	if err != nil {
		return nil, err
	}

	err = v.db.GetContext(ctx, &vendorAdmin, query, args...)
	if err != nil {

		if err == sql.ErrNoRows {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return &vendorAdmin, nil
}
func (v *VendorAdminDB) GetVendorAdmins(ctx context.Context, vendorID uuid.UUID) ([]VendorAdminUser, error) {
	vendorinfo := []VendorAdminUser{}
	query, args, err := QB.Select("vendor_admins.user_id, vendor_admins.vendor_id, users.email").
		From("vendor_admins").
		Join("users ON vendor_admins.user_id = users.id").
		Where(squirrel.Eq{"vendor_admins.vendor_id": vendorID}).
		ToSql()
	if err != nil {
		return nil, err
	}

	err = v.db.SelectContext(ctx, &vendorinfo, query, args...)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrRecordNotFound
		}
		return nil, err
	}
	return vendorinfo, nil
}

// UpdateVendorAdmin updates an existing vendor admin record in the database.
func (v *VendorAdminDB) UpdateVendorAdmin(ctx context.Context, vendor VendorAdmin) (*VendorAdmin, error) {
	query, args, err := QB.Update("vendor_admins").
		Set("user_id", vendor.UserID).
		Set("vendor_id", vendor.VendorID).
		Where(squirrel.Eq{"user_id": vendor.UserID, "vendor_id": vendor.VendorID}).
		Suffix("RETURNING user_id, vendor_id").
		ToSql()
	if err != nil {
		return nil, err
	}

	err = v.db.QueryRowxContext(ctx, query, args...).StructScan(&vendor)
	if err != nil {
		return nil, fmt.Errorf("error while updating vendor admin: %v", err)
	}
	return &vendor, nil
}

// DeleteVendorAdmin deletes a vendor admin record by user_id and vendor_id.
func (v *VendorAdminDB) DeleteVendorAdmin(ctx context.Context, userID, vendorID uuid.UUID) error {
	query, args, err := QB.Delete("vendor_admins").
		Where(squirrel.Eq{"user_id": userID, "vendor_id": vendorID}).
		ToSql()
	if err != nil {
		return err
	}

	result, err := v.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("error while deleting vendor admin: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return fmt.Errorf("vendor admin not found")
	}
	return nil
}
