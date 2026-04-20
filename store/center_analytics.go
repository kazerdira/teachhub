package store

import (
	"context"
	"time"
)

// CenterDashboardStats provides enriched stats for the center owner dashboard.
type CenterDashboardStats struct {
	TeacherCount      int
	ActiveStudents    int
	SessionsThisMonth int
}

func (s *Store) GetCenterDashboardStats(ctx context.Context, centerID int) (*CenterDashboardStats, error) {
	st := &CenterDashboardStats{}

	s.DB.QueryRow(ctx,
		`SELECT COUNT(*) FROM admin WHERE center_id=$1 AND active=true AND role='teacher'`, centerID).Scan(&st.TeacherCount)

	s.DB.QueryRow(ctx, `
		SELECT COUNT(DISTINCT cs.student_id) FROM classroom_student cs
		JOIN classroom cl ON cl.id=cs.classroom_id
		JOIN admin a ON a.id=cl.admin_id
		WHERE a.center_id=$1 AND cs.status='approved'`, centerID).Scan(&st.ActiveStudents)

	s.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM live_session ls
		JOIN classroom cl ON cl.id=ls.classroom_id
		JOIN admin a ON a.id=cl.admin_id
		WHERE a.center_id=$1 AND ls.created_at >= date_trunc('month', NOW())`, centerID).Scan(&st.SessionsThisMonth)

	return st, nil
}

// TeacherPerformanceRow shows per-teacher metrics for the center dashboard.
type TeacherPerformanceRow struct {
	TeacherID         int
	Username          string
	DisplayName       string
	Email             string
	ClassroomCount    int
	StudentCount      int
	AvgQuizPct        float64
	SessionsThisMonth int
	LastActive        *time.Time
}

func (s *Store) GetCenterTeacherPerformance(ctx context.Context, centerID int) ([]TeacherPerformanceRow, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT a.id, a.username, COALESCE(a.display_name,''), a.email,
			(SELECT COUNT(*) FROM classroom WHERE admin_id=a.id),
			(SELECT COUNT(DISTINCT cs.student_id) FROM classroom_student cs
			 JOIN classroom cl ON cl.id=cs.classroom_id
			 WHERE cl.admin_id=a.id AND cs.status='approved'),
			COALESCE((SELECT AVG(CASE WHEN qa.max_score>0 THEN qa.score*100.0/qa.max_score ELSE 0 END)
			 FROM quiz_attempt qa
			 JOIN quiz q ON q.id=qa.quiz_id
			 JOIN classroom cl ON cl.id=q.classroom_id
			 WHERE cl.admin_id=a.id AND qa.finished_at IS NOT NULL), 0),
			COALESCE((SELECT COUNT(*) FROM live_session ls
			 JOIN classroom cl ON cl.id=ls.classroom_id
			 WHERE cl.admin_id=a.id AND ls.created_at >= date_trunc('month', NOW())), 0),
			a.last_login_at
		FROM admin a
		WHERE a.center_id=$1 AND a.role='teacher' AND a.active=true
		ORDER BY a.username`, centerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []TeacherPerformanceRow
	for rows.Next() {
		var r TeacherPerformanceRow
		if err := rows.Scan(&r.TeacherID, &r.Username, &r.DisplayName, &r.Email, &r.ClassroomCount, &r.StudentCount,
			&r.AvgQuizPct, &r.SessionsThisMonth, &r.LastActive); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, nil
}
