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

// CenterActivityEvent is one row in the center owner's activity feed.
// Kind: "submission", "quiz", "join_request", "live_session", "student_joined", "teacher_added"
type CenterActivityEvent struct {
	At            time.Time
	Kind          string
	Actor         string // student or teacher name
	Detail        string // e.g. quiz title, assignment title, classroom name
	TeacherName   string
	ClassroomName string
}

// GetCenterActivityFeed returns the most recent N events across the center.
// Uses UNION ALL on existing tables — no new infra required.
func (s *Store) GetCenterActivityFeed(ctx context.Context, centerID, limit int) ([]CenterActivityEvent, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.DB.Query(ctx, `
		SELECT at, kind, actor, detail, teacher_name, classroom_name FROM (
			-- Submissions
			SELECT s.submitted_at AS at, 'submission'::text AS kind,
				st.name AS actor,
				asg.title AS detail,
				COALESCE(a.display_name, a.username) AS teacher_name,
				cl.name AS classroom_name
			FROM submission s
			JOIN student st ON st.id = s.student_id
			JOIN assignment asg ON asg.id = s.assignment_id
			JOIN classroom cl ON cl.id = asg.classroom_id
			JOIN admin a ON a.id = cl.admin_id
			WHERE a.center_id = $1

			UNION ALL

			-- Finished quiz attempts
			SELECT qa.finished_at AS at, 'quiz'::text AS kind,
				st.name AS actor,
				q.title AS detail,
				COALESCE(a.display_name, a.username) AS teacher_name,
				cl.name AS classroom_name
			FROM quiz_attempt qa
			JOIN student st ON st.id = qa.student_id
			JOIN quiz q ON q.id = qa.quiz_id
			JOIN classroom cl ON cl.id = q.classroom_id
			JOIN admin a ON a.id = cl.admin_id
			WHERE a.center_id = $1 AND qa.finished_at IS NOT NULL

			UNION ALL

			-- Join requests
			SELECT jr.created_at AS at, 'join_request'::text AS kind,
				jr.full_name AS actor,
				COALESCE(cl.name, '') AS detail,
				COALESCE(a.display_name, a.username) AS teacher_name,
				COALESCE(cl.name, '') AS classroom_name
			FROM join_request jr
			JOIN admin a ON a.id = jr.teacher_id
			LEFT JOIN classroom cl ON cl.id = jr.classroom_id
			WHERE a.center_id = $1

			UNION ALL

			-- Live sessions started
			SELECT ls.created_at AS at, 'live_session'::text AS kind,
				COALESCE(a.display_name, a.username) AS actor,
				cl.name AS detail,
				COALESCE(a.display_name, a.username) AS teacher_name,
				cl.name AS classroom_name
			FROM live_session ls
			JOIN classroom cl ON cl.id = ls.classroom_id
			JOIN admin a ON a.id = cl.admin_id
			WHERE a.center_id = $1

			UNION ALL

			-- Students approved into a classroom
			SELECT cs.joined_at AS at, 'student_joined'::text AS kind,
				st.name AS actor,
				cl.name AS detail,
				COALESCE(a.display_name, a.username) AS teacher_name,
				cl.name AS classroom_name
			FROM classroom_student cs
			JOIN student st ON st.id = cs.student_id
			JOIN classroom cl ON cl.id = cs.classroom_id
			JOIN admin a ON a.id = cl.admin_id
			WHERE a.center_id = $1 AND cs.status = 'approved'
		) feed
		ORDER BY at DESC
		LIMIT $2`, centerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []CenterActivityEvent
	for rows.Next() {
		var e CenterActivityEvent
		if err := rows.Scan(&e.At, &e.Kind, &e.Actor, &e.Detail, &e.TeacherName, &e.ClassroomName); err != nil {
			return nil, err
		}
		list = append(list, e)
	}
	return list, nil
}

// CenterTodayPulse — small counters scoped to today (UTC server day).
type CenterTodayPulse struct {
	SubmissionsToday   int
	QuizAttemptsToday  int
	LiveSessionsToday  int
	NewStudentsToday   int
}

func (s *Store) GetCenterTodayPulse(ctx context.Context, centerID int) (*CenterTodayPulse, error) {
	p := &CenterTodayPulse{}
	s.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM submission s
		JOIN assignment asg ON asg.id = s.assignment_id
		JOIN classroom cl ON cl.id = asg.classroom_id
		JOIN admin a ON a.id = cl.admin_id
		WHERE a.center_id=$1 AND s.submitted_at >= date_trunc('day', NOW())`, centerID).Scan(&p.SubmissionsToday)
	s.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM quiz_attempt qa
		JOIN quiz q ON q.id = qa.quiz_id
		JOIN classroom cl ON cl.id = q.classroom_id
		JOIN admin a ON a.id = cl.admin_id
		WHERE a.center_id=$1 AND qa.finished_at >= date_trunc('day', NOW())`, centerID).Scan(&p.QuizAttemptsToday)
	s.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM live_session ls
		JOIN classroom cl ON cl.id = ls.classroom_id
		JOIN admin a ON a.id = cl.admin_id
		WHERE a.center_id=$1 AND ls.created_at >= date_trunc('day', NOW())`, centerID).Scan(&p.LiveSessionsToday)
	s.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM classroom_student cs
		JOIN classroom cl ON cl.id = cs.classroom_id
		JOIN admin a ON a.id = cl.admin_id
		WHERE a.center_id=$1 AND cs.status='approved' AND cs.joined_at >= date_trunc('day', NOW())`, centerID).Scan(&p.NewStudentsToday)
	return p, nil
}
