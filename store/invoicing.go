package store

import (
	"context"
	"time"
)

// GenerateCenterMonthlyInvoice generates (or returns existing) the invoice for a center
// for the given period month.  Counts teachers who are:
//   - role='teacher', billable_from <= period_end
//   - AND (active=true OR deactivated_at >= period_start)
//
// Idempotent: UNIQUE(center_id, period_month) → ON CONFLICT DO NOTHING.
func (s *Store) GenerateCenterMonthlyInvoice(ctx context.Context, centerID int, periodMonth time.Time) (*CenterInvoice, error) {
	periodStart := time.Date(periodMonth.Year(), periodMonth.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0)

	// Count billable teachers for this period
	var teacherCount int
	err := s.DB.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM admin
		WHERE center_id = $1
		  AND role = 'teacher'
		  AND billable_from IS NOT NULL
		  AND billable_from <= $3
		  AND (active = true OR deactivated_at >= $2)`,
		centerID, periodStart, periodEnd).Scan(&teacherCount)
	if err != nil {
		return nil, err
	}

	// Get pricing snapshot from center
	var pricePerTeacher float64
	var currency string
	err = s.DB.QueryRow(ctx,
		`SELECT price_per_teacher, currency FROM center WHERE id=$1`, centerID).
		Scan(&pricePerTeacher, &currency)
	if err != nil {
		return nil, err
	}

	totalAmount := float64(teacherCount) * pricePerTeacher

	_, err = s.DB.Exec(ctx, `
		INSERT INTO center_invoice
		    (center_id, period_month, teacher_count, price_per_teacher, currency, total_amount)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (center_id, period_month) DO NOTHING`,
		centerID, periodStart, teacherCount, pricePerTeacher, currency, totalAmount)
	if err != nil {
		return nil, err
	}

	// Return the row (may be the existing one from a prior call)
	inv := &CenterInvoice{}
	err = s.DB.QueryRow(ctx, `
		SELECT id, center_id, period_month, teacher_count, price_per_teacher,
		       currency, total_amount, status, paid_at, paid_method, paid_reference, generated_at
		FROM center_invoice
		WHERE center_id=$1 AND period_month=$2`,
		centerID, periodStart).
		Scan(&inv.ID, &inv.CenterID, &inv.PeriodMonth, &inv.TeacherCount,
			&inv.PricePerTeacher, &inv.Currency, &inv.TotalAmount,
			&inv.Status, &inv.PaidAt, &inv.PaidMethod, &inv.PaidReference, &inv.GeneratedAt)
	if err != nil {
		return nil, err
	}
	return inv, nil
}

// HasUnpaidCenterInvoice returns true if the center has at least one unpaid invoice.
func (s *Store) HasUnpaidCenterInvoice(ctx context.Context, centerID int) bool {
	var count int
	s.DB.QueryRow(ctx,
		`SELECT COUNT(*) FROM center_invoice WHERE center_id=$1 AND status='unpaid'`,
		centerID).Scan(&count)
	return count > 0
}

// ListCenterInvoices returns all invoices for a center, newest first.
func (s *Store) ListCenterInvoices(ctx context.Context, centerID int) ([]CenterInvoice, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, center_id, period_month, teacher_count, price_per_teacher,
		       currency, total_amount, status, paid_at, paid_method, paid_reference, generated_at
		FROM center_invoice
		WHERE center_id=$1
		ORDER BY period_month DESC`, centerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []CenterInvoice
	for rows.Next() {
		var inv CenterInvoice
		if err := rows.Scan(&inv.ID, &inv.CenterID, &inv.PeriodMonth, &inv.TeacherCount,
			&inv.PricePerTeacher, &inv.Currency, &inv.TotalAmount,
			&inv.Status, &inv.PaidAt, &inv.PaidMethod, &inv.PaidReference, &inv.GeneratedAt); err != nil {
			return nil, err
		}
		list = append(list, inv)
	}
	return list, nil
}

// MarkCenterInvoicePaid marks an invoice as paid.
func (s *Store) MarkCenterInvoicePaid(ctx context.Context, invoiceID, centerID int, method, reference string) error {
	_, err := s.DB.Exec(ctx,
		`UPDATE center_invoice SET status='paid', paid_at=NOW(), paid_method=$3, paid_reference=$4
		 WHERE id=$1 AND center_id=$2`,
		invoiceID, centerID, method, reference)
	return err
}

// CancelCenterInvoice cancels an invoice.
func (s *Store) CancelCenterInvoice(ctx context.Context, invoiceID, centerID int) error {
	_, err := s.DB.Exec(ctx,
		`UPDATE center_invoice SET status='cancelled' WHERE id=$1 AND center_id=$2`,
		invoiceID, centerID)
	return err
}
