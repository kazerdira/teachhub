package store

import (
	"context"
	"time"
)

// ─── Center model ───────────────────────────────────────

type Center struct {
	ID                 int
	Name               string
	OwnerAdminID       *int
	Address            string
	City               string
	Country            string
	Phone              string
	Email              string
	LogoPath           string
	SubscriptionStatus string // trial, active, expired, suspended, cancelled
	SubscriptionStart  *time.Time
	SubscriptionEnd    *time.Time
	SeatCount          int
	PricePerSeat       float64
	TrialEndsAt        *time.Time
	CreatedAt          time.Time
}

// CenterTeacher is a lightweight view used in the center teacher list.
type CenterTeacher struct {
	ID             int
	Username       string
	DisplayName    string
	Email          string
	Phone          string
	Role           string
	Active         bool
	ClassroomCount int
	StudentCount   int
	LastLoginAt    *time.Time
	CreatedAt      time.Time
}

// ─── Center CRUD ────────────────────────────────────────

func (s *Store) CreateCenter(ctx context.Context, name, email string, ownerID int) (int, error) {
	var id int
	err := s.DB.QueryRow(ctx,
		`INSERT INTO center (name, email, subscription_status, trial_ends_at)
		 VALUES ($1, $2, 'trial', NOW() + INTERVAL '30 days')
		 RETURNING id`, name, email).Scan(&id)
	return id, err
}

func (s *Store) GetCenter(ctx context.Context, id int) (*Center, error) {
	c := &Center{}
	err := s.DB.QueryRow(ctx,
		`SELECT id, name, owner_admin_id, address, city, country, phone, email, logo_path,
		        subscription_status, subscription_start, subscription_end,
		        seat_count, price_per_seat, trial_ends_at, created_at
		 FROM center WHERE id=$1`, id).
		Scan(&c.ID, &c.Name, &c.OwnerAdminID, &c.Address, &c.City, &c.Country,
			&c.Phone, &c.Email, &c.LogoPath,
			&c.SubscriptionStatus, &c.SubscriptionStart, &c.SubscriptionEnd,
			&c.SeatCount, &c.PricePerSeat, &c.TrialEndsAt, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Store) UpdateCenter(ctx context.Context, id int, name, address, city, phone, email string) error {
	_, err := s.DB.Exec(ctx,
		`UPDATE center SET name=$1, address=$2, city=$3, phone=$4, email=$5 WHERE id=$6`,
		name, address, city, phone, email, id)
	return err
}

// ─── Center teachers ────────────────────────────────────

func (s *Store) ListCenterTeachers(ctx context.Context, centerID int) ([]CenterTeacher, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT a.id, a.username, COALESCE(a.display_name,''), a.email, COALESCE(a.phone,''), a.role, a.active,
		       COALESCE((SELECT COUNT(*) FROM classroom WHERE admin_id = a.id), 0),
		       COALESCE((SELECT COUNT(DISTINCT cs.student_id)
		                 FROM classroom_student cs
		                 JOIN classroom c ON c.id = cs.classroom_id
		                 WHERE c.admin_id = a.id AND cs.status='approved'), 0),
		       a.last_login_at, a.created_at
		FROM admin a
		WHERE a.center_id = $1 AND a.role = 'teacher'
		ORDER BY a.created_at ASC`, centerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []CenterTeacher
	for rows.Next() {
		var t CenterTeacher
		if err := rows.Scan(&t.ID, &t.Username, &t.DisplayName, &t.Email, &t.Phone, &t.Role, &t.Active,
			&t.ClassroomCount, &t.StudentCount, &t.LastLoginAt, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, nil
}

func (s *Store) CountCenterTeachers(ctx context.Context, centerID int) (int, error) {
	var count int
	err := s.DB.QueryRow(ctx,
		`SELECT COUNT(*) FROM admin WHERE center_id=$1 AND active=true AND role='teacher'`, centerID).Scan(&count)
	return count, err
}

func (s *Store) CreateOwnerAdmin(ctx context.Context, centerID int, username, hashedPassword, plaintextPassword, email, phone, schoolName string, applicationID int, displayName string) (int, error) {
	var id int
	err := s.DB.QueryRow(ctx,
		`INSERT INTO admin (username, password, pending_password, email, phone, school_name,
		                     subscription_status, subscription_start, created_by_platform, application_id,
		                     role, center_id, active, display_name)
		 VALUES ($1, $2, $3, $4, $5, $6, 'active', NOW(), true, $7, 'owner', $8, true, $9)
		 RETURNING id`,
		username, hashedPassword, plaintextPassword, email, phone, schoolName, applicationID, centerID, displayName).Scan(&id)
	return id, err
}

func (s *Store) CreateTeacherInCenter(ctx context.Context, centerID int, username, hashedPassword, plaintextPassword, email, phone, displayName string) (int, error) {
	var id int
	err := s.DB.QueryRow(ctx,
		`INSERT INTO admin (username, password, pending_password, email, phone,
		                     subscription_status, subscription_start, created_by_platform,
		                     role, center_id, active, display_name)
		 VALUES ($1, $2, $3, $4, $5, 'active', NOW(), true, 'teacher', $6, true, $7)
		 RETURNING id`,
		username, hashedPassword, plaintextPassword, email, phone, centerID, displayName).Scan(&id)
	return id, err
}

func (s *Store) DeactivateTeacher(ctx context.Context, teacherID int) error {
	_, err := s.DB.Exec(ctx, `UPDATE admin SET active=false WHERE id=$1`, teacherID)
	return err
}

func (s *Store) ActivateTeacher(ctx context.Context, teacherID int) error {
	_, err := s.DB.Exec(ctx, `UPDATE admin SET active=true WHERE id=$1`, teacherID)
	return err
}

// ─── Center stats for dashboard ─────────────────────────

type CenterStats struct {
	TeacherCount int
	StudentCount int
	ClassCount   int
	SessionCount int
	ActiveSeats  int
	SeatCount    int
}

func (s *Store) GetCenterStats(ctx context.Context, centerID int) (*CenterStats, error) {
	st := &CenterStats{}

	// Active teachers (exclude owner — owner doesn't occupy a seat)
	s.DB.QueryRow(ctx,
		`SELECT COUNT(*) FROM admin WHERE center_id=$1 AND active=true AND role='teacher'`, centerID).Scan(&st.ActiveSeats)
	st.TeacherCount = st.ActiveSeats

	// Total students across center teachers
	s.DB.QueryRow(ctx,
		`SELECT COUNT(DISTINCT cs.student_id)
		 FROM classroom_student cs
		 JOIN classroom c ON c.id = cs.classroom_id
		 JOIN admin a ON a.id = c.admin_id
		 WHERE a.center_id=$1 AND cs.status='approved'`, centerID).Scan(&st.StudentCount)

	// Total classrooms
	s.DB.QueryRow(ctx,
		`SELECT COUNT(*)
		 FROM classroom c
		 JOIN admin a ON a.id = c.admin_id
		 WHERE a.center_id=$1`, centerID).Scan(&st.ClassCount)

	// Live sessions
	s.DB.QueryRow(ctx,
		`SELECT COUNT(*)
		 FROM live_session ls
		 JOIN classroom c ON c.id = ls.classroom_id
		 JOIN admin a ON a.id = c.admin_id
		 WHERE a.center_id=$1`, centerID).Scan(&st.SessionCount)

	return st, nil
}

// ─── Platform: list all centers ─────────────────────────

type CenterListItem struct {
	ID                 int
	Name               string
	OwnerUsername      string
	Email              string
	SubscriptionStatus string
	SeatCount          int
	ActiveTeachers     int
	StudentCount       int
	CreatedAt          time.Time
}

func (s *Store) ListCenters(ctx context.Context) ([]CenterListItem, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT c.id, c.name, COALESCE(a.username,''), c.email, c.subscription_status,
		       c.seat_count,
		       COALESCE((SELECT COUNT(*) FROM admin WHERE center_id=c.id AND active=true), 0),
		       COALESCE((SELECT COUNT(DISTINCT cs.student_id)
		                 FROM classroom_student cs
		                 JOIN classroom cl ON cl.id = cs.classroom_id
		                 JOIN admin t ON t.id = cl.admin_id
		                 WHERE t.center_id=c.id AND cs.status='approved'), 0),
		       c.created_at
		FROM center c
		LEFT JOIN admin a ON a.id = c.owner_admin_id
		ORDER BY c.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []CenterListItem
	for rows.Next() {
		var ci CenterListItem
		if err := rows.Scan(&ci.ID, &ci.Name, &ci.OwnerUsername, &ci.Email, &ci.SubscriptionStatus,
			&ci.SeatCount, &ci.ActiveTeachers, &ci.StudentCount, &ci.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, ci)
	}
	return list, nil
}

func (s *Store) UpdateCenterSubscription(ctx context.Context, centerID int, status string, start, end *time.Time) error {
	_, err := s.DB.Exec(ctx,
		`UPDATE center SET subscription_status=$1, subscription_start=$2, subscription_end=$3 WHERE id=$4`,
		status, start, end, centerID)
	return err
}

func (s *Store) UpdateCenterSeats(ctx context.Context, centerID, seats int, price float64) error {
	_, err := s.DB.Exec(ctx,
		`UPDATE center SET seat_count=$1, price_per_seat=$2 WHERE id=$3`,
		seats, price, centerID)
	return err
}

// ─── Center students ────────────────────────────────────

// CenterStudent is a view of a student at the center level with classroom assignments.
type CenterStudent struct {
	ID             int
	Name           string
	Email          string
	Phone          string
	ClassroomCount int
	ClassroomNames string // comma-separated
	CreatedAt      time.Time
	LastLoginAt    *time.Time
}

// ListCenterStudents returns all students belonging to this center.
func (s *Store) ListCenterStudents(ctx context.Context, centerID int) ([]CenterStudent, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT s.id, s.name, COALESCE(s.email,''), COALESCE(s.phone,''),
		       COALESCE((SELECT COUNT(DISTINCT cs.classroom_id)
		                 FROM classroom_student cs
		                 JOIN classroom c ON c.id = cs.classroom_id
		                 JOIN admin a ON a.id = c.admin_id
		                 WHERE cs.student_id = s.id AND a.center_id = $1 AND cs.status='approved'), 0),
		       COALESCE((SELECT string_agg(DISTINCT c.name, ', ')
		                 FROM classroom_student cs
		                 JOIN classroom c ON c.id = cs.classroom_id
		                 JOIN admin a ON a.id = c.admin_id
		                 WHERE cs.student_id = s.id AND a.center_id = $1 AND cs.status='approved'), ''),
		       s.created_at, s.last_login_at
		FROM student s
		WHERE s.center_id = $1
		ORDER BY s.name ASC`, centerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []CenterStudent
	for rows.Next() {
		var st CenterStudent
		if err := rows.Scan(&st.ID, &st.Name, &st.Email, &st.Phone,
			&st.ClassroomCount, &st.ClassroomNames, &st.CreatedAt, &st.LastLoginAt); err != nil {
			return nil, err
		}
		list = append(list, st)
	}
	return list, nil
}

// CountCenterStudents returns the total number of students at a center.
func (s *Store) CountCenterStudents(ctx context.Context, centerID int) (int, error) {
	var count int
	err := s.DB.QueryRow(ctx, `SELECT COUNT(*) FROM student WHERE center_id=$1`, centerID).Scan(&count)
	return count, err
}

// CreateCenterStudent creates a student at the center level (not yet assigned to any classroom).
func (s *Store) CreateCenterStudent(ctx context.Context, centerID int, name, email, phone string) (int, error) {
	var id int
	err := s.DB.QueryRow(ctx,
		`INSERT INTO student (name, email, phone, center_id) VALUES ($1, $2, $3, $4) RETURNING id`,
		name, email, phone, centerID).Scan(&id)
	return id, err
}

// AssignStudentToClassroom assigns a center student to a classroom (auto-approved).
func (s *Store) AssignStudentToClassroom(ctx context.Context, studentID, classroomID int) error {
	_, err := s.DB.Exec(ctx,
		`INSERT INTO classroom_student (classroom_id, student_id, status) VALUES ($1, $2, 'approved')
		 ON CONFLICT (classroom_id, student_id) DO UPDATE SET status='approved'`,
		classroomID, studentID)
	return err
}

// ListCenterClassrooms returns all classrooms in a center (across all teachers).
func (s *Store) ListCenterClassrooms(ctx context.Context, centerID int) ([]Classroom, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT c.id, c.name, c.join_code, COALESCE(c.subject,''), COALESCE(c.level,''),
		       a.id, COALESCE(a.display_name, a.username),
		       COALESCE((SELECT COUNT(*) FROM classroom_student cs WHERE cs.classroom_id=c.id AND cs.status='approved'),0),
		       c.created_at
		FROM classroom c
		JOIN admin a ON a.id = c.admin_id
		WHERE a.center_id = $1
		ORDER BY a.display_name, c.name`, centerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	type classWithTeacher struct {
		Classroom
		TeacherID   int
		TeacherName string
	}
	var list []Classroom
	for rows.Next() {
		var cl Classroom
		var teacherID int
		var teacherName string
		if err := rows.Scan(&cl.ID, &cl.Name, &cl.JoinCode, &cl.Subject, &cl.Level,
			&teacherID, &teacherName, &cl.StudentCount, &cl.CreatedAt); err != nil {
			return nil, err
		}
		// Store teacher name in the Subject field temporarily for display
		// (we'll use a proper struct in the template)
		cl.AdminID = teacherID
		cl.TeacherName = teacherName // teacher display name for center context
		list = append(list, cl)
	}
	return list, nil
}
