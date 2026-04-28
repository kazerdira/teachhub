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
	PricePerTeacher    float64
	Currency           string
	BillingMode        string
	TrialEndsAt        *time.Time
	CreatedAt          time.Time
}

// CenterTeacher is a lightweight view used in the center teacher list.
type CenterTeacher struct {
	ID              int
	Username        string
	DisplayName     string
	Email           string
	Phone           string
	Role            string
	Active          bool
	ClassroomCount  int
	StudentCount    int
	LastLoginAt     *time.Time
	CreatedAt       time.Time
	BillableFrom    *time.Time
	DeactivatedAt   *time.Time
	PendingPassword *string
}

// ─── Center CRUD ────────────────────────────────────────

func (s *Store) CreateCenter(ctx context.Context, name, email string, ownerID int, currency string, pricePerTeacher float64) (int, error) {
	return s.CreateCenterWithCountry(ctx, name, email, ownerID, currency, pricePerTeacher, "")
}

// CreateCenterWithCountry inserts a new center, persisting the ISO country code so that
// downstream features (subjects, levels, billing) reflect the right locale from day one.
func (s *Store) CreateCenterWithCountry(ctx context.Context, name, email string, ownerID int, currency string, pricePerTeacher float64, country string) (int, error) {
	if country == "" {
		country = "DZ"
	}
	var id int
	err := s.DB.QueryRow(ctx,
		`INSERT INTO center (name, email, country, subscription_status, trial_ends_at, currency, price_per_teacher)
		 VALUES ($1, $2, $3, 'trial', NOW() + INTERVAL '30 days', $4, $5)
		 RETURNING id`, name, email, country, currency, pricePerTeacher).Scan(&id)
	return id, err
}

func (s *Store) GetCenter(ctx context.Context, id int) (*Center, error) {
	c := &Center{}
	err := s.DB.QueryRow(ctx,
		`SELECT id, name, owner_admin_id, address, city, country, phone, email, logo_path,
		        subscription_status, subscription_start, subscription_end,
		        price_per_teacher, currency, billing_mode, trial_ends_at, created_at
		 FROM center WHERE id=$1`, id).
		Scan(&c.ID, &c.Name, &c.OwnerAdminID, &c.Address, &c.City, &c.Country,
			&c.Phone, &c.Email, &c.LogoPath,
			&c.SubscriptionStatus, &c.SubscriptionStart, &c.SubscriptionEnd,
			&c.PricePerTeacher, &c.Currency, &c.BillingMode, &c.TrialEndsAt, &c.CreatedAt)
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
		       a.last_login_at, a.created_at,
		       a.billable_from, a.deactivated_at, a.pending_password
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
			&t.ClassroomCount, &t.StudentCount, &t.LastLoginAt, &t.CreatedAt,
			&t.BillableFrom, &t.DeactivatedAt, &t.PendingPassword); err != nil {
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
	_, err := s.DB.Exec(ctx, `UPDATE admin SET active=false, deactivated_at=NOW() WHERE id=$1`, teacherID)
	return err
}

func (s *Store) ActivateTeacher(ctx context.Context, teacherID int) error {
	_, err := s.DB.Exec(ctx, `UPDATE admin SET active=true, deactivated_at=NULL WHERE id=$1`, teacherID)
	return err
}

// ─── Center stats for dashboard ─────────────────────────

type CenterStats struct {
	TeacherCount int
	StudentCount int
	ClassCount   int
	SessionCount int
}

func (s *Store) GetCenterStats(ctx context.Context, centerID int) (*CenterStats, error) {
	st := &CenterStats{}

	// Active teachers (exclude owner)
	s.DB.QueryRow(ctx,
		`SELECT COUNT(*) FROM admin WHERE center_id=$1 AND active=true AND role='teacher'`, centerID).Scan(&st.TeacherCount)

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
	ActiveTeachers     int
	StudentCount       int
	CreatedAt          time.Time
}

func (s *Store) ListCenters(ctx context.Context) ([]CenterListItem, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT c.id, c.name, COALESCE(a.username,''), c.email, c.subscription_status,
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
			&ci.ActiveTeachers, &ci.StudentCount, &ci.CreatedAt); err != nil {
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

func (s *Store) UpdateCenterPricing(ctx context.Context, centerID int, pricePerTeacher float64, currency string) error {
	_, err := s.DB.Exec(ctx,
		`UPDATE center SET price_per_teacher=$1, currency=$2 WHERE id=$3`,
		pricePerTeacher, currency, centerID)
	return err
}

func (s *Store) SetBillableFrom(ctx context.Context, adminID int, t time.Time) error {
	_, err := s.DB.Exec(ctx,
		`UPDATE admin SET billable_from=$2 WHERE id=$1 AND billable_from IS NULL`,
		adminID, t)
	return err
}

func (s *Store) SetDeactivatedAt(ctx context.Context, adminID int, t *time.Time) error {
	_, err := s.DB.Exec(ctx,
		`UPDATE admin SET deactivated_at=$2 WHERE id=$1`,
		adminID, t)
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

// ─── Center student detail ──────────────────────────────

// CenterStudentMembership — one (student, classroom) row inside a center.
type CenterStudentMembership struct {
	ClassroomID   int
	ClassroomName string
	Subject       string
	Level         string
	TeacherName   string
	Status        string
	JoinedAt      time.Time
}

// CenterStudentDetail — full student profile + memberships + activity counts.
type CenterStudentDetail struct {
	ID               int
	Name             string
	Email            string
	Phone            string
	CreatedAt        time.Time
	LastLoginAt      *time.Time
	LastLoginIP      string
	Memberships      []CenterStudentMembership
	QuizAttempts     int
	SessionsAttended int
	Submissions      int
}

func (s *Store) GetCenterStudentDetail(ctx context.Context, studentID, centerID int) (*CenterStudentDetail, error) {
	d := &CenterStudentDetail{}
	err := s.DB.QueryRow(ctx, `
		SELECT s.id, s.name, COALESCE(s.email,''), COALESCE(s.phone,''),
		       s.created_at, s.last_login_at, COALESCE(s.last_login_ip,'')
		FROM student s WHERE s.id=$1 AND s.center_id=$2`, studentID, centerID).
		Scan(&d.ID, &d.Name, &d.Email, &d.Phone, &d.CreatedAt, &d.LastLoginAt, &d.LastLoginIP)
	if err != nil {
		return nil, err
	}

	rows, err := s.DB.Query(ctx, `
		SELECT c.id, c.name, COALESCE(c.subject,''), COALESCE(c.level,''),
		       COALESCE(a.display_name, a.username),
		       cs.status, cs.joined_at
		FROM classroom_student cs
		JOIN classroom c ON c.id = cs.classroom_id
		JOIN admin a ON a.id = c.admin_id
		WHERE cs.student_id=$1 AND a.center_id=$2
		ORDER BY cs.joined_at DESC`, studentID, centerID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var m CenterStudentMembership
			if err := rows.Scan(&m.ClassroomID, &m.ClassroomName, &m.Subject, &m.Level,
				&m.TeacherName, &m.Status, &m.JoinedAt); err == nil {
				d.Memberships = append(d.Memberships, m)
			}
		}
	}

	_ = s.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM quiz_attempt qa
		JOIN quiz q ON q.id = qa.quiz_id
		JOIN classroom c ON c.id = q.classroom_id
		JOIN admin a ON a.id = c.admin_id
		WHERE qa.student_id=$1 AND a.center_id=$2`, studentID, centerID).Scan(&d.QuizAttempts)

	_ = s.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM live_attendance la
		JOIN live_session ls ON ls.id = la.live_session_id
		JOIN classroom c ON c.id = ls.classroom_id
		JOIN admin a ON a.id = c.admin_id
		WHERE la.student_id=$1 AND a.center_id=$2`, studentID, centerID).Scan(&d.SessionsAttended)

	_ = s.DB.QueryRow(ctx, `
		SELECT COUNT(*) FROM submission sb
		JOIN assignment ag ON ag.id = sb.assignment_id
		JOIN classroom c ON c.id = ag.classroom_id
		JOIN admin a ON a.id = c.admin_id
		WHERE sb.student_id=$1 AND a.center_id=$2`, studentID, centerID).Scan(&d.Submissions)

	return d, nil
}

// UpdateCenterStudent updates basic profile fields for a student belonging to the center.
func (s *Store) UpdateCenterStudent(ctx context.Context, studentID, centerID int, name, email, phone string) error {
	_, err := s.DB.Exec(ctx,
		`UPDATE student SET name=$3, email=$4, phone=$5 WHERE id=$1 AND center_id=$2`,
		studentID, centerID, name, email, phone)
	return err
}

// DeleteCenterStudent fully removes a student from the center (cascades all memberships, attempts, etc.).
func (s *Store) DeleteCenterStudent(ctx context.Context, studentID, centerID int) error {
	_, err := s.DB.Exec(ctx,
		`DELETE FROM student WHERE id=$1 AND center_id=$2`, studentID, centerID)
	return err
}

// ─── Center classroom detail ────────────────────────────

// CenterClassroomStudent — student row shown inside a classroom detail page.
type CenterClassroomStudent struct {
	ID       int
	Name     string
	Email    string
	Phone    string
	Status   string
	JoinedAt time.Time
}

// CenterClassroomDetail — classroom + roster + quizzes + counts.
type CenterClassroomDetail struct {
	Classroom
	Students        []CenterClassroomStudent
	Quizzes         []Quiz
	SessionCount    int
	AssignmentCount int
}

func (s *Store) GetCenterClassroomDetail(ctx context.Context, classroomID, centerID int) (*CenterClassroomDetail, error) {
	d := &CenterClassroomDetail{}
	err := s.DB.QueryRow(ctx, `
		SELECT c.id, c.name, c.join_code, COALESCE(c.subject,''), COALESCE(c.level,''),
		       a.id, COALESCE(a.display_name, a.username),
		       c.created_at
		FROM classroom c
		JOIN admin a ON a.id = c.admin_id
		WHERE c.id=$1 AND a.center_id=$2`, classroomID, centerID).
		Scan(&d.Classroom.ID, &d.Classroom.Name, &d.Classroom.JoinCode,
			&d.Classroom.Subject, &d.Classroom.Level, &d.Classroom.AdminID,
			&d.Classroom.TeacherName, &d.Classroom.CreatedAt)
	if err != nil {
		return nil, err
	}

	rows, err := s.DB.Query(ctx, `
		SELECT s.id, s.name, COALESCE(s.email,''), COALESCE(s.phone,''), cs.status, cs.joined_at
		FROM classroom_student cs
		JOIN student s ON s.id = cs.student_id
		WHERE cs.classroom_id=$1
		ORDER BY cs.status, s.name`, classroomID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var st CenterClassroomStudent
			if err := rows.Scan(&st.ID, &st.Name, &st.Email, &st.Phone, &st.Status, &st.JoinedAt); err == nil {
				d.Students = append(d.Students, st)
				if st.Status == "approved" {
					d.Classroom.StudentCount++
				}
			}
		}
	}

	qrows, err := s.DB.Query(ctx, `
		SELECT id, title, COALESCE(description,''), published, deadline, created_at
		FROM quiz WHERE classroom_id=$1 ORDER BY created_at DESC`, classroomID)
	if err == nil {
		defer qrows.Close()
		for qrows.Next() {
			var q Quiz
			q.ClassroomID = classroomID
			if err := qrows.Scan(&q.ID, &q.Title, &q.Description, &q.Published, &q.Deadline, &q.CreatedAt); err == nil {
				d.Quizzes = append(d.Quizzes, q)
			}
		}
	}

	_ = s.DB.QueryRow(ctx, `SELECT COUNT(*) FROM live_session WHERE classroom_id=$1`, classroomID).Scan(&d.SessionCount)
	_ = s.DB.QueryRow(ctx, `SELECT COUNT(*) FROM assignment WHERE classroom_id=$1`, classroomID).Scan(&d.AssignmentCount)

	return d, nil
}


