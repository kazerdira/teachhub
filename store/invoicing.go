package store

import (
	"context"
	"time"
)

// GenerateMonthlyInvoices generates invoices for a center for a given month.
// Idempotent: UNIQUE constraint on (classroom_id, student_id, period_month).
func (s *Store) GenerateMonthlyInvoices(ctx context.Context, centerID int, period time.Time) (int, error) {
	periodStart := time.Date(period.Year(), period.Month(), 1, 0, 0, 0, 0, time.UTC)
	periodEnd := periodStart.AddDate(0, 1, 0)

	tag, err := s.DB.Exec(ctx, `
		INSERT INTO student_invoice (center_id, classroom_id, student_id, period_month,
		                             sessions_attended, rate_per_session, total_amount)
		SELECT $1,
		       cl.id,
		       la.student_id,
		       $2::date,
		       COUNT(DISTINCT la.live_session_id),
		       cl.session_rate,
		       COUNT(DISTINCT la.live_session_id) * cl.session_rate
		FROM live_attendance la
		JOIN live_session ls ON ls.id = la.live_session_id
		JOIN classroom cl ON cl.id = ls.classroom_id
		JOIN admin a ON a.id = cl.admin_id
		WHERE a.center_id = $1
		  AND cl.billing_enabled = true
		  AND cl.session_rate > 0
		  AND ls.created_at >= $2
		  AND ls.created_at < $3
		  AND ls.duration_minutes >= 5
		GROUP BY cl.id, la.student_id, cl.session_rate
		HAVING COUNT(DISTINCT la.live_session_id) > 0
		ON CONFLICT (classroom_id, student_id, period_month) DO NOTHING`,
		centerID, periodStart, periodEnd)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

// ListCenterInvoices returns invoices for a center filtered by period and status.
func (s *Store) ListCenterInvoices(ctx context.Context, centerID int, period time.Time, status string) ([]StudentInvoice, error) {
	periodStart := time.Date(period.Year(), period.Month(), 1, 0, 0, 0, 0, time.UTC)
	q := `SELECT si.id, si.center_id, si.classroom_id, si.student_id, s.name, cl.name,
	             si.period_month, si.sessions_attended, si.rate_per_session, si.total_amount,
	             si.status, si.paid_at, si.paid_method, si.notes, si.generated_at
	      FROM student_invoice si
	      JOIN student s ON s.id = si.student_id
	      JOIN classroom cl ON cl.id = si.classroom_id
	      WHERE si.center_id = $1 AND si.period_month = $2`
	args := []interface{}{centerID, periodStart}
	if status != "" && status != "all" {
		q += ` AND si.status = $3`
		args = append(args, status)
	}
	q += ` ORDER BY si.status, s.name`
	rows, err := s.DB.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []StudentInvoice
	for rows.Next() {
		var inv StudentInvoice
		if err := rows.Scan(&inv.ID, &inv.CenterID, &inv.ClassroomID, &inv.StudentID, &inv.StudentName, &inv.ClassroomName,
			&inv.PeriodMonth, &inv.SessionsAttended, &inv.RatePerSession, &inv.TotalAmount,
			&inv.Status, &inv.PaidAt, &inv.PaidMethod, &inv.Notes, &inv.GeneratedAt); err != nil {
			return nil, err
		}
		list = append(list, inv)
	}
	return list, nil
}

// MarkInvoicePaid marks an invoice as paid with a method.
func (s *Store) MarkInvoicePaid(ctx context.Context, invoiceID, centerID int, method string) error {
	_, err := s.DB.Exec(ctx,
		`UPDATE student_invoice SET status='paid', paid_at=NOW(), paid_method=$3
		 WHERE id=$1 AND center_id=$2`, invoiceID, centerID, method)
	return err
}

// CancelInvoice marks an invoice as cancelled.
func (s *Store) CancelInvoice(ctx context.Context, invoiceID, centerID int) error {
	_, err := s.DB.Exec(ctx,
		`UPDATE student_invoice SET status='cancelled'
		 WHERE id=$1 AND center_id=$2`, invoiceID, centerID)
	return err
}

// GetStudentInvoices returns recent invoices for a student in a classroom (for parent report).
func (s *Store) GetStudentInvoices(ctx context.Context, studentID, classroomID int) ([]StudentInvoice, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, center_id, classroom_id, student_id, '', '',
		       period_month, sessions_attended, rate_per_session, total_amount,
		       status, paid_at, paid_method, notes, generated_at
		FROM student_invoice
		WHERE student_id=$1 AND classroom_id=$2
		ORDER BY period_month DESC
		LIMIT 6`, studentID, classroomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []StudentInvoice
	for rows.Next() {
		var inv StudentInvoice
		if err := rows.Scan(&inv.ID, &inv.CenterID, &inv.ClassroomID, &inv.StudentID, &inv.StudentName, &inv.ClassroomName,
			&inv.PeriodMonth, &inv.SessionsAttended, &inv.RatePerSession, &inv.TotalAmount,
			&inv.Status, &inv.PaidAt, &inv.PaidMethod, &inv.Notes, &inv.GeneratedAt); err != nil {
			return nil, err
		}
		list = append(list, inv)
	}
	return list, nil
}

// GetCenterBillingSummary returns totals for a center in a given period.
func (s *Store) GetCenterBillingSummary(ctx context.Context, centerID int, period time.Time) (totalAmount float64, paidAmount float64, unpaidCount int, err error) {
	periodStart := time.Date(period.Year(), period.Month(), 1, 0, 0, 0, 0, time.UTC)
	s.DB.QueryRow(ctx,
		`SELECT COALESCE(SUM(total_amount),0), COALESCE(SUM(CASE WHEN status='paid' THEN total_amount ELSE 0 END),0),
		        COALESCE(SUM(CASE WHEN status='unpaid' THEN 1 ELSE 0 END),0)
		 FROM student_invoice WHERE center_id=$1 AND period_month=$2`,
		centerID, periodStart).Scan(&totalAmount, &paidAmount, &unpaidCount)
	return
}

// LogParentView logs a parent report view.
func (s *Store) LogParentView(ctx context.Context, parentCode, ip string) {
	s.DB.Exec(ctx, `INSERT INTO parent_view_log (parent_code, ip) VALUES ($1, $2)`, parentCode, ip)
}

// GetParentViewsWeek returns how many parent views happened this week for a center.
func (s *Store) GetParentViewsWeek(ctx context.Context, centerID int) int {
	var count int
	s.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM parent_view_log pvl
		JOIN classroom_student cs ON cs.parent_code = pvl.parent_code
		JOIN classroom cl ON cl.id = cs.classroom_id
		JOIN admin a ON a.id = cl.admin_id
		WHERE a.center_id=$1 AND pvl.viewed_at >= NOW() - INTERVAL '7 days'`, centerID).Scan(&count)
	return count
}
