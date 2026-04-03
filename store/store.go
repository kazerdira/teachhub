package store

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	DB *pgxpool.Pool
}

func New(db *pgxpool.Pool) *Store {
	return &Store{DB: db}
}

// ─── Models ─────────────────────────────────────────────

type Admin struct {
	ID                 int
	Username           string
	Password           string
	Email              string
	SchoolName         string
	SubscriptionStatus string // active, expired, suspended
	SubscriptionStart  *time.Time
	SubscriptionEnd    *time.Time
	CreatedByPlatform  bool
	ApplicationID      *int
	PendingPassword    *string
}

type TeacherListItem struct {
	ID                 int
	Username           string
	Email              string
	SchoolName         string
	SubscriptionStatus string
	SubscriptionStart  *time.Time
	SubscriptionEnd    *time.Time
	CreatedAt          time.Time
	ClassroomCount     int
	StudentCount       int
}

type PlatformAdmin struct {
	ID           int
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

type TeacherApplication struct {
	ID         int
	FullName   string
	Email      string
	Phone      string
	SchoolName string
	Wilaya     string
	Message    string
	Status     string // pending, approved, rejected, contacted
	AdminNotes string
	CreatedAt  time.Time
	ReviewedAt *time.Time
}

type Payment struct {
	ID         int
	TeacherID  int
	Amount     float64
	Method     string // cash, ccp, baridi_mob, other
	Reference  string
	Notes      string
	RecordedAt time.Time
}

type Classroom struct {
	ID         int
	Name       string
	JoinCode   string
	TeacherPic string
	CreatedAt  time.Time
	// computed
	StudentCount  int
	ResourceCount int
	QuizCount     int
	PendingCount  int
}

type Student struct {
	ID           int
	Name         string
	Email        string
	CreatedAt    time.Time
	MemberStatus string // approved, pending, rejected — only set in classroom context
	ParentCode   string // per-classroom secret link for parent reports
}

type AllowedStudent struct {
	ID          int
	ClassroomID int
	Email       string
	Name        string
	CreatedAt   time.Time
}

type Category struct {
	ID          int
	ClassroomID int
	Name        string
	SortOrder   int
}

type Resource struct {
	ID           int
	ClassroomID  int
	CategoryID   *int
	CategoryName string
	Title        string
	Description  string
	FilePath     string
	FileType     string
	ExternalURL  string
	FileSize     int64
	CreatedAt    time.Time
}

type Assignment struct {
	ID              int
	ClassroomID     int
	Title           string
	Description     string
	Deadline        *time.Time
	ResponseType    string // file, text, both
	MaxChars        int
	MaxFileSize     int64 // bytes, default 10MB
	MaxGrade        float64
	FilePath        string
	FileName        string
	CreatedAt       time.Time
	SubmissionCount int
	GradedCount     int
}

type Submission struct {
	ID           int
	AssignmentID int
	StudentID    int
	StudentName  string
	FilePath     string
	FileName     string
	FileSize     int64
	TextContent  string
	Status       string
	Feedback     string
	Grade        *float64
	MaxGrade     *float64
	GradedAt     *time.Time
	SubmittedAt  time.Time
}

type Quiz struct {
	ID               int
	ClassroomID      int
	Title            string
	Description      string
	Published        bool
	Deadline         *time.Time
	TimeLimitMinutes int
	MaxAttempts      int
	CreatedAt        time.Time
	// computed
	QuestionCount int
	AttemptCount  int
}

type QuizQuestion struct {
	ID            int
	QuizID        int
	SortOrder     int
	QuestionType  string
	Content       string
	Options       []string
	CorrectAnswer string
	Points        int
}

type QuizAttempt struct {
	ID          int
	QuizID      int
	StudentID   int
	StudentName string
	Answers     map[string]string
	FileAnswers map[string]map[string]string // question_id -> {file_path, file_name}
	Score       *int
	MaxScore    *int
	Reviewed    bool
	StartedAt   time.Time
	FinishedAt  *time.Time
}

type LiveSession struct {
	ID              int
	ClassroomID     int
	RoomName        string
	Active          bool
	EndedAt         *time.Time
	DurationMinutes *int
	CreatedAt       time.Time
	// computed
	AttendeeCount int
}

// ─── Admin ──────────────────────────────────────────────

func (s *Store) GetAdmin(ctx context.Context, username string) (*Admin, error) {
	a := &Admin{}
	err := s.DB.QueryRow(ctx,
		`SELECT id, username, password, email, school_name, subscription_status,
		        subscription_start, subscription_end, created_by_platform, application_id, pending_password
		 FROM admin WHERE username=$1`, username).
		Scan(&a.ID, &a.Username, &a.Password, &a.Email, &a.SchoolName,
			&a.SubscriptionStatus, &a.SubscriptionStart, &a.SubscriptionEnd,
			&a.CreatedByPlatform, &a.ApplicationID, &a.PendingPassword)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Store) CreateAdmin(ctx context.Context, username, hashedPassword string) error {
	_, err := s.DB.Exec(ctx,
		`INSERT INTO admin (username, password) VALUES ($1, $2) ON CONFLICT (username) DO NOTHING`,
		username, hashedPassword)
	return err
}

// ─── Classrooms ─────────────────────────────────────────

func (s *Store) ListClassrooms(ctx context.Context, adminID int) ([]Classroom, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT c.id, c.name, c.join_code, c.teacher_pic, c.created_at,
			(SELECT COUNT(*) FROM classroom_student WHERE classroom_id=c.id AND status='approved') AS student_count,
			(SELECT COUNT(*) FROM resource WHERE classroom_id=c.id) AS resource_count,
			(SELECT COUNT(*) FROM quiz WHERE classroom_id=c.id) AS quiz_count,
			(SELECT COUNT(*) FROM classroom_student WHERE classroom_id=c.id AND status='pending') AS pending_count
		FROM classroom c WHERE c.admin_id=$1 ORDER BY c.created_at DESC`, adminID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Classroom
	for rows.Next() {
		var c Classroom
		if err := rows.Scan(&c.ID, &c.Name, &c.JoinCode, &c.TeacherPic, &c.CreatedAt, &c.StudentCount, &c.ResourceCount, &c.QuizCount, &c.PendingCount); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, nil
}

func (s *Store) GetClassroom(ctx context.Context, id int) (*Classroom, error) {
	c := &Classroom{}
	err := s.DB.QueryRow(ctx, `SELECT id, name, join_code, teacher_pic, created_at FROM classroom WHERE id=$1`, id).
		Scan(&c.ID, &c.Name, &c.JoinCode, &c.TeacherPic, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Store) GetClassroomForAdmin(ctx context.Context, id, adminID int) (*Classroom, error) {
	c := &Classroom{}
	err := s.DB.QueryRow(ctx, `SELECT id, name, join_code, teacher_pic, created_at FROM classroom WHERE id=$1 AND admin_id=$2`, id, adminID).
		Scan(&c.ID, &c.Name, &c.JoinCode, &c.TeacherPic, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Store) GetClassroomByCode(ctx context.Context, code string) (*Classroom, error) {
	c := &Classroom{}
	err := s.DB.QueryRow(ctx, `SELECT id, name, join_code, teacher_pic, created_at FROM classroom WHERE join_code=$1`, code).
		Scan(&c.ID, &c.Name, &c.JoinCode, &c.TeacherPic, &c.CreatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (s *Store) CreateClassroom(ctx context.Context, name string, adminID int) (int, error) {
	var id int
	err := s.DB.QueryRow(ctx, `INSERT INTO classroom (name, admin_id) VALUES ($1, $2) RETURNING id`, name, adminID).Scan(&id)
	return id, err
}

func (s *Store) DeleteClassroom(ctx context.Context, id, adminID int) error {
	_, err := s.DB.Exec(ctx, `DELETE FROM classroom WHERE id=$1 AND admin_id=$2`, id, adminID)
	return err
}

func (s *Store) RegenerateJoinCode(ctx context.Context, id, adminID int) (string, error) {
	var code string
	err := s.DB.QueryRow(ctx,
		`UPDATE classroom SET join_code=encode(gen_random_bytes(4),'hex') WHERE id=$1 AND admin_id=$2 RETURNING join_code`, id, adminID).
		Scan(&code)
	return code, err
}

func (s *Store) SetClassroomTeacherPic(ctx context.Context, classroomID, adminID int, picPath string) error {
	_, err := s.DB.Exec(ctx, `UPDATE classroom SET teacher_pic=$1 WHERE id=$2 AND admin_id=$3`, picPath, classroomID, adminID)
	return err
}

// ─── Students ───────────────────────────────────────────

func (s *Store) ListClassroomStudents(ctx context.Context, classroomID int) ([]Student, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT s.id, s.name, COALESCE(s.email,''), s.created_at, cs.status, COALESCE(cs.parent_code,'')
		FROM student s JOIN classroom_student cs ON s.id=cs.student_id
		WHERE cs.classroom_id=$1 ORDER BY cs.status, s.name`, classroomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Student
	for rows.Next() {
		var st Student
		if err := rows.Scan(&st.ID, &st.Name, &st.Email, &st.CreatedAt, &st.MemberStatus, &st.ParentCode); err != nil {
			return nil, err
		}
		list = append(list, st)
	}
	return list, nil
}

func (s *Store) GetStudent(ctx context.Context, id int) (*Student, error) {
	st := &Student{}
	err := s.DB.QueryRow(ctx, `SELECT id, name, COALESCE(email,''), created_at FROM student WHERE id=$1`, id).
		Scan(&st.ID, &st.Name, &st.Email, &st.CreatedAt)
	if err != nil {
		return nil, err
	}
	return st, nil
}

func (s *Store) CreateStudentAndJoin(ctx context.Context, name, email string, classroomID int) (int, error) {
	return s.CreateStudentAndJoinWithStatus(ctx, name, email, classroomID, "approved")
}

func (s *Store) CreateStudentAndJoinWithStatus(ctx context.Context, name, email string, classroomID int, status string) (int, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var id int
	err = tx.QueryRow(ctx, `INSERT INTO student (name, email) VALUES ($1, $2) RETURNING id`, name, email).Scan(&id)
	if err != nil {
		return 0, err
	}
	_, err = tx.Exec(ctx, `INSERT INTO classroom_student (classroom_id, student_id, status) VALUES ($1, $2, $3)`, classroomID, id, status)
	if err != nil {
		return 0, err
	}
	return id, tx.Commit(ctx)
}

func (s *Store) CreateStudentAndJoinExisting(ctx context.Context, studentID, classroomID int) error {
	_, err := s.DB.Exec(ctx,
		`INSERT INTO classroom_student (classroom_id, student_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
		classroomID, studentID)
	return err
}

func (s *Store) CreateStudentAndJoinExistingWithStatus(ctx context.Context, studentID, classroomID int, status string) error {
	_, err := s.DB.Exec(ctx,
		`INSERT INTO classroom_student (classroom_id, student_id, status) VALUES ($1, $2, $3) ON CONFLICT DO NOTHING`,
		classroomID, studentID, status)
	return err
}

func (s *Store) IsStudentInClassroom(ctx context.Context, studentID, classroomID int) (bool, error) {
	var exists bool
	err := s.DB.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM classroom_student WHERE classroom_id=$1 AND student_id=$2 AND status='approved')`,
		classroomID, studentID).Scan(&exists)
	return exists, err
}

func (s *Store) IsStudentMemberOfClassroom(ctx context.Context, studentID, classroomID int) (bool, string, error) {
	var status string
	err := s.DB.QueryRow(ctx,
		`SELECT status FROM classroom_student WHERE classroom_id=$1 AND student_id=$2`,
		classroomID, studentID).Scan(&status)
	if err != nil {
		return false, "", err
	}
	return true, status, nil
}

func (s *Store) GetStudentClassrooms(ctx context.Context, studentID int) ([]Classroom, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT c.id, c.name, c.join_code, c.created_at
		FROM classroom c JOIN classroom_student cs ON c.id=cs.classroom_id
		WHERE cs.student_id=$1 ORDER BY cs.joined_at DESC`, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Classroom
	for rows.Next() {
		var c Classroom
		if err := rows.Scan(&c.ID, &c.Name, &c.JoinCode, &c.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, nil
}

func (s *Store) RemoveStudentFromClassroom(ctx context.Context, studentID, classroomID int) error {
	_, err := s.DB.Exec(ctx, `DELETE FROM classroom_student WHERE classroom_id=$1 AND student_id=$2`, classroomID, studentID)
	return err
}

// ─── Parent Report ──────────────────────────────────────

// ParentReportData holds everything needed to render the parent progress page.
type ParentReportData struct {
	StudentID     int
	StudentName   string
	ClassroomID   int
	ClassroomName string
	TeacherName   string
}

// GetStudentByParentCode looks up a student+classroom from the secret parent_code.
func (s *Store) GetStudentByParentCode(ctx context.Context, code string) (*ParentReportData, error) {
	d := &ParentReportData{}
	err := s.DB.QueryRow(ctx, `
		SELECT cs.student_id, s.name, cs.classroom_id, c.name, a.username
		FROM classroom_student cs
		JOIN student s ON s.id = cs.student_id
		JOIN classroom c ON c.id = cs.classroom_id
		JOIN admin a ON a.id = c.admin_id
		WHERE cs.parent_code = $1 AND cs.status = 'approved'`, code).
		Scan(&d.StudentID, &d.StudentName, &d.ClassroomID, &d.ClassroomName, &d.TeacherName)
	if err != nil {
		return nil, err
	}
	return d, nil
}

// RegenerateParentCode creates a new parent_code for a student in a classroom.
func (s *Store) RegenerateParentCode(ctx context.Context, classroomID, studentID int) (string, error) {
	var code string
	err := s.DB.QueryRow(ctx,
		`UPDATE classroom_student SET parent_code = encode(gen_random_bytes(6), 'hex')
		 WHERE classroom_id = $1 AND student_id = $2 RETURNING parent_code`,
		classroomID, studentID).Scan(&code)
	return code, err
}

// GetParentCode returns the current parent_code for a student in a classroom.
func (s *Store) GetParentCode(ctx context.Context, classroomID, studentID int) (string, error) {
	var code string
	err := s.DB.QueryRow(ctx,
		`SELECT COALESCE(parent_code, '') FROM classroom_student WHERE classroom_id = $1 AND student_id = $2`,
		classroomID, studentID).Scan(&code)
	return code, err
}

// ─── Allowed Students ───────────────────────────────────

func (s *Store) IsEmailAllowed(ctx context.Context, email string, classroomID int) (bool, error) {
	var exists bool
	err := s.DB.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM allowed_student WHERE classroom_id=$1 AND LOWER(email)=LOWER($2))`,
		classroomID, email).Scan(&exists)
	return exists, err
}

func (s *Store) ListAllowedStudents(ctx context.Context, classroomID int) ([]AllowedStudent, error) {
	rows, err := s.DB.Query(ctx,
		`SELECT id, classroom_id, email, COALESCE(name,''), created_at FROM allowed_student WHERE classroom_id=$1 ORDER BY email`,
		classroomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []AllowedStudent
	for rows.Next() {
		var a AllowedStudent
		if err := rows.Scan(&a.ID, &a.ClassroomID, &a.Email, &a.Name, &a.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, nil
}

func (s *Store) AddAllowedStudent(ctx context.Context, classroomID int, email, name string) error {
	_, err := s.DB.Exec(ctx,
		`INSERT INTO allowed_student (classroom_id, email, name) VALUES ($1, LOWER($2), $3) ON CONFLICT (classroom_id, email) DO NOTHING`,
		classroomID, email, name)
	return err
}

func (s *Store) DeleteAllowedStudent(ctx context.Context, id, classroomID int) error {
	_, err := s.DB.Exec(ctx, `DELETE FROM allowed_student WHERE id=$1 AND classroom_id=$2`, id, classroomID)
	return err
}

func (s *Store) ApproveStudent(ctx context.Context, studentID, classroomID int) error {
	_, err := s.DB.Exec(ctx,
		`UPDATE classroom_student SET status='approved' WHERE classroom_id=$1 AND student_id=$2`,
		classroomID, studentID)
	return err
}

func (s *Store) RejectStudent(ctx context.Context, studentID, classroomID int) error {
	_, err := s.DB.Exec(ctx,
		`DELETE FROM classroom_student WHERE classroom_id=$1 AND student_id=$2`,
		classroomID, studentID)
	return err
}

func (s *Store) CountPendingStudents(ctx context.Context, classroomID int) (int, error) {
	var count int
	err := s.DB.QueryRow(ctx,
		`SELECT COUNT(*) FROM classroom_student WHERE classroom_id=$1 AND status='pending'`,
		classroomID).Scan(&count)
	return count, err
}

// ─── Categories ─────────────────────────────────────────

func (s *Store) ListCategories(ctx context.Context, classroomID int) ([]Category, error) {
	rows, err := s.DB.Query(ctx,
		`SELECT id, classroom_id, name, sort_order FROM category WHERE classroom_id=$1 ORDER BY sort_order, name`,
		classroomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.ClassroomID, &c.Name, &c.SortOrder); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, nil
}

func (s *Store) CreateCategory(ctx context.Context, classroomID int, name string) (int, error) {
	var id int
	err := s.DB.QueryRow(ctx,
		`INSERT INTO category (classroom_id, name) VALUES ($1, $2) RETURNING id`,
		classroomID, name).Scan(&id)
	return id, err
}

func (s *Store) DeleteCategory(ctx context.Context, id, classroomID int) error {
	_, err := s.DB.Exec(ctx, `DELETE FROM category WHERE id=$1 AND classroom_id=$2`, id, classroomID)
	return err
}

// ─── Resources ──────────────────────────────────────────

func (s *Store) ListResources(ctx context.Context, classroomID int) ([]Resource, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT r.id, r.classroom_id, r.category_id, COALESCE(c.name,'Uncategorized'),
			r.title, r.description, COALESCE(r.file_path,''), COALESCE(r.file_type,''),
			COALESCE(r.external_url,''), r.file_size, r.created_at
		FROM resource r LEFT JOIN category c ON r.category_id=c.id
		WHERE r.classroom_id=$1 ORDER BY COALESCE(c.sort_order,999), c.name, r.created_at DESC`,
		classroomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Resource
	for rows.Next() {
		var r Resource
		if err := rows.Scan(&r.ID, &r.ClassroomID, &r.CategoryID, &r.CategoryName,
			&r.Title, &r.Description, &r.FilePath, &r.FileType,
			&r.ExternalURL, &r.FileSize, &r.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, nil
}

func (s *Store) CreateResource(ctx context.Context, classroomID int, categoryID *int, title, desc, filePath, fileType, externalURL string, fileSize int64) (int, error) {
	var id int
	err := s.DB.QueryRow(ctx, `
		INSERT INTO resource (classroom_id, category_id, title, description, file_path, file_type, external_url, file_size)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING id`,
		classroomID, categoryID, title, desc, filePath, fileType, externalURL, fileSize).Scan(&id)
	return id, err
}

func (s *Store) GetResource(ctx context.Context, id int) (*Resource, error) {
	r := &Resource{}
	err := s.DB.QueryRow(ctx, `
		SELECT id, classroom_id, category_id, '', title, description,
			COALESCE(file_path,''), COALESCE(file_type,''), COALESCE(external_url,''), file_size, created_at
		FROM resource WHERE id=$1`, id).
		Scan(&r.ID, &r.ClassroomID, &r.CategoryID, &r.CategoryName,
			&r.Title, &r.Description, &r.FilePath, &r.FileType, &r.ExternalURL, &r.FileSize, &r.CreatedAt)
	return r, err
}

func (s *Store) DeleteResource(ctx context.Context, id, classroomID int) error {
	_, err := s.DB.Exec(ctx, `DELETE FROM resource WHERE id=$1 AND classroom_id=$2`, id, classroomID)
	return err
}

// ─── Assignments ────────────────────────────────────────

func (s *Store) ListAssignments(ctx context.Context, classroomID int) ([]Assignment, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT a.id, a.classroom_id, a.title, a.description, a.deadline,
			a.response_type, a.max_chars, a.max_file_size, a.max_grade,
			COALESCE(a.file_path,''), COALESCE(a.file_name,''), a.created_at,
			(SELECT COUNT(*) FROM submission WHERE assignment_id=a.id) AS sub_count,
			(SELECT COUNT(*) FROM submission WHERE assignment_id=a.id AND grade IS NOT NULL) AS graded_count
		FROM assignment a WHERE a.classroom_id=$1 ORDER BY a.created_at DESC`, classroomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Assignment
	for rows.Next() {
		var a Assignment
		if err := rows.Scan(&a.ID, &a.ClassroomID, &a.Title, &a.Description, &a.Deadline,
			&a.ResponseType, &a.MaxChars, &a.MaxFileSize, &a.MaxGrade,
			&a.FilePath, &a.FileName, &a.CreatedAt, &a.SubmissionCount, &a.GradedCount); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, nil
}

func (s *Store) GetAssignment(ctx context.Context, id int) (*Assignment, error) {
	a := &Assignment{}
	err := s.DB.QueryRow(ctx, `SELECT id, classroom_id, title, description, deadline, response_type, max_chars, max_file_size, max_grade, COALESCE(file_path,''), COALESCE(file_name,''), created_at FROM assignment WHERE id=$1`, id).
		Scan(&a.ID, &a.ClassroomID, &a.Title, &a.Description, &a.Deadline, &a.ResponseType, &a.MaxChars, &a.MaxFileSize, &a.MaxGrade, &a.FilePath, &a.FileName, &a.CreatedAt)
	return a, err
}

func (s *Store) CreateAssignment(ctx context.Context, classroomID int, title, desc string, deadline *time.Time, responseType string, maxChars int, maxFileSize int64, maxGrade float64, filePath, fileName string) (int, error) {
	if responseType == "" {
		responseType = "file"
	}
	if maxFileSize <= 0 {
		maxFileSize = 10 * 1024 * 1024 // 10MB default
	}
	if maxGrade <= 0 {
		maxGrade = 20
	}
	var id int
	err := s.DB.QueryRow(ctx,
		`INSERT INTO assignment (classroom_id, title, description, deadline, response_type, max_chars, max_file_size, max_grade, file_path, file_name) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) RETURNING id`,
		classroomID, title, desc, deadline, responseType, maxChars, maxFileSize, maxGrade, filePath, fileName).Scan(&id)
	return id, err
}

func (s *Store) DeleteAssignment(ctx context.Context, id, classroomID int) error {
	_, err := s.DB.Exec(ctx, `DELETE FROM assignment WHERE id=$1 AND classroom_id=$2`, id, classroomID)
	return err
}

func (s *Store) UpdateAssignment(ctx context.Context, id, classroomID int, title, desc string, deadline *time.Time, responseType string, maxChars int, maxFileSize int64, maxGrade float64, filePath, fileName string) error {
	_, err := s.DB.Exec(ctx,
		`UPDATE assignment SET title=$3, description=$4, deadline=$5, response_type=$6, max_chars=$7, max_file_size=$8, max_grade=$9, file_path=$10, file_name=$11
		 WHERE id=$1 AND classroom_id=$2`,
		id, classroomID, title, desc, deadline, responseType, maxChars, maxFileSize, maxGrade, filePath, fileName)
	return err
}

// ─── Submissions ────────────────────────────────────────

func (s *Store) ListSubmissions(ctx context.Context, assignmentID int) ([]Submission, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT sub.id, sub.assignment_id, sub.student_id, s.name,
			COALESCE(sub.file_path,''), COALESCE(sub.file_name,''), sub.file_size,
			COALESCE(sub.text_content,''), sub.status, sub.feedback,
			sub.grade, sub.max_grade, sub.graded_at, sub.submitted_at
		FROM submission sub JOIN student s ON sub.student_id=s.id
		WHERE sub.assignment_id=$1 ORDER BY sub.submitted_at DESC`, assignmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Submission
	for rows.Next() {
		var sub Submission
		if err := rows.Scan(&sub.ID, &sub.AssignmentID, &sub.StudentID, &sub.StudentName,
			&sub.FilePath, &sub.FileName, &sub.FileSize,
			&sub.TextContent, &sub.Status, &sub.Feedback,
			&sub.Grade, &sub.MaxGrade, &sub.GradedAt, &sub.SubmittedAt); err != nil {
			return nil, err
		}
		list = append(list, sub)
	}
	return list, nil
}

func (s *Store) ListSubmissionsPaged(ctx context.Context, assignmentID, limit, offset int) ([]Submission, int, error) {
	var total int
	s.DB.QueryRow(ctx, `SELECT COUNT(*) FROM submission WHERE assignment_id=$1`, assignmentID).Scan(&total)

	rows, err := s.DB.Query(ctx, `
		SELECT sub.id, sub.assignment_id, sub.student_id, s.name,
			COALESCE(sub.file_path,''), COALESCE(sub.file_name,''), sub.file_size,
			COALESCE(sub.text_content,''), sub.status, sub.feedback,
			sub.grade, sub.max_grade, sub.graded_at, sub.submitted_at
		FROM submission sub JOIN student s ON sub.student_id=s.id
		WHERE sub.assignment_id=$1 ORDER BY sub.submitted_at DESC
		LIMIT $2 OFFSET $3`, assignmentID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []Submission
	for rows.Next() {
		var sub Submission
		if err := rows.Scan(&sub.ID, &sub.AssignmentID, &sub.StudentID, &sub.StudentName,
			&sub.FilePath, &sub.FileName, &sub.FileSize,
			&sub.TextContent, &sub.Status, &sub.Feedback,
			&sub.Grade, &sub.MaxGrade, &sub.GradedAt, &sub.SubmittedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, sub)
	}
	return list, total, nil
}

func (s *Store) GetStudentSubmissions(ctx context.Context, assignmentID, studentID int) ([]Submission, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, assignment_id, student_id, '',
			COALESCE(file_path,''), COALESCE(file_name,''), file_size,
			COALESCE(text_content,''), status, feedback,
			grade, max_grade, graded_at, submitted_at
		FROM submission WHERE assignment_id=$1 AND student_id=$2 ORDER BY submitted_at DESC`,
		assignmentID, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Submission
	for rows.Next() {
		var sub Submission
		if err := rows.Scan(&sub.ID, &sub.AssignmentID, &sub.StudentID, &sub.StudentName,
			&sub.FilePath, &sub.FileName, &sub.FileSize,
			&sub.TextContent, &sub.Status, &sub.Feedback,
			&sub.Grade, &sub.MaxGrade, &sub.GradedAt, &sub.SubmittedAt); err != nil {
			return nil, err
		}
		list = append(list, sub)
	}
	return list, nil
}

func (s *Store) CreateSubmission(ctx context.Context, assignmentID, studentID int, filePath, fileName string, fileSize int64, textContent string) (int, error) {
	var id int
	err := s.DB.QueryRow(ctx,
		`INSERT INTO submission (assignment_id, student_id, file_path, file_name, file_size, text_content) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		assignmentID, studentID, filePath, fileName, fileSize, textContent).Scan(&id)
	return id, err
}

func (s *Store) UpdateSubmissionStatus(ctx context.Context, id int, classroomID int, status, feedback string, grade *float64, maxGrade *float64) error {
	if grade != nil {
		_, err := s.DB.Exec(ctx, `UPDATE submission SET status=$2, feedback=$3, grade=$4, max_grade=$5, graded_at=NOW()
			WHERE id=$1 AND assignment_id IN (SELECT id FROM assignment WHERE classroom_id=$6)`,
			id, status, feedback, *grade, *maxGrade, classroomID)
		return err
	}
	_, err := s.DB.Exec(ctx, `UPDATE submission SET status=$2, feedback=$3
		WHERE id=$1 AND assignment_id IN (SELECT id FROM assignment WHERE classroom_id=$4)`,
		id, status, feedback, classroomID)
	return err
}

// ─── Quizzes ────────────────────────────────────────────

func (s *Store) ListQuizzes(ctx context.Context, classroomID int) ([]Quiz, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT q.id, q.classroom_id, q.title, q.description, q.published, q.deadline, q.time_limit_minutes, q.max_attempts, q.created_at,
			(SELECT COUNT(*) FROM quiz_question WHERE quiz_id=q.id) AS q_count,
			(SELECT COUNT(DISTINCT student_id) FROM quiz_attempt WHERE quiz_id=q.id) AS a_count
		FROM quiz q WHERE q.classroom_id=$1 ORDER BY q.created_at DESC`, classroomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Quiz
	for rows.Next() {
		var q Quiz
		if err := rows.Scan(&q.ID, &q.ClassroomID, &q.Title, &q.Description, &q.Published, &q.Deadline, &q.TimeLimitMinutes, &q.MaxAttempts, &q.CreatedAt, &q.QuestionCount, &q.AttemptCount); err != nil {
			return nil, err
		}
		list = append(list, q)
	}
	return list, nil
}

func (s *Store) GetQuiz(ctx context.Context, id int) (*Quiz, error) {
	q := &Quiz{}
	err := s.DB.QueryRow(ctx,
		`SELECT id, classroom_id, title, description, published, deadline, time_limit_minutes, max_attempts, created_at FROM quiz WHERE id=$1`, id).
		Scan(&q.ID, &q.ClassroomID, &q.Title, &q.Description, &q.Published, &q.Deadline, &q.TimeLimitMinutes, &q.MaxAttempts, &q.CreatedAt)
	return q, err
}

func (s *Store) CreateQuiz(ctx context.Context, classroomID int, title, desc string, deadline *time.Time, timeLimitMin, maxAttempts int) (int, error) {
	if maxAttempts <= 0 {
		maxAttempts = 1
	}
	var id int
	err := s.DB.QueryRow(ctx,
		`INSERT INTO quiz (classroom_id, title, description, deadline, time_limit_minutes, max_attempts) VALUES ($1,$2,$3,$4,$5,$6) RETURNING id`,
		classroomID, title, desc, deadline, timeLimitMin, maxAttempts).Scan(&id)
	return id, err
}

func (s *Store) UpdateQuiz(ctx context.Context, id int, title, desc string, published bool, deadline *time.Time, timeLimitMin, maxAttempts int) error {
	_, err := s.DB.Exec(ctx, `UPDATE quiz SET title=$2, description=$3, published=$4, deadline=$5, time_limit_minutes=$6, max_attempts=$7 WHERE id=$1`,
		id, title, desc, published, deadline, timeLimitMin, maxAttempts)
	return err
}

func (s *Store) DeleteQuiz(ctx context.Context, id, classroomID int) error {
	_, err := s.DB.Exec(ctx, `DELETE FROM quiz WHERE id=$1 AND classroom_id=$2`, id, classroomID)
	return err
}

// ─── Quiz Questions ─────────────────────────────────────

func (s *Store) ListQuizQuestions(ctx context.Context, quizID int) ([]QuizQuestion, error) {
	rows, err := s.DB.Query(ctx,
		`SELECT id, quiz_id, sort_order, question_type, content, options, COALESCE(correct_answer,''), points
		FROM quiz_question WHERE quiz_id=$1 ORDER BY sort_order`, quizID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []QuizQuestion
	for rows.Next() {
		var q QuizQuestion
		var optJSON []byte
		if err := rows.Scan(&q.ID, &q.QuizID, &q.SortOrder, &q.QuestionType, &q.Content, &optJSON, &q.CorrectAnswer, &q.Points); err != nil {
			return nil, err
		}
		if optJSON != nil {
			json.Unmarshal(optJSON, &q.Options)
		}
		list = append(list, q)
	}
	return list, nil
}

func (s *Store) CreateQuizQuestion(ctx context.Context, quizID, sortOrder int, qType, content string, options []string, correctAnswer string, points int) (int, error) {
	optJSON, _ := json.Marshal(options)
	var id int
	err := s.DB.QueryRow(ctx, `
		INSERT INTO quiz_question (quiz_id, sort_order, question_type, content, options, correct_answer, points)
		VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING id`,
		quizID, sortOrder, qType, content, optJSON, correctAnswer, points).Scan(&id)
	return id, err
}

func (s *Store) UpdateQuizQuestion(ctx context.Context, id int, qType, content string, options []string, correctAnswer string, points int) error {
	optJSON, _ := json.Marshal(options)
	_, err := s.DB.Exec(ctx, `
		UPDATE quiz_question SET question_type=$2, content=$3, options=$4, correct_answer=$5, points=$6 WHERE id=$1`,
		id, qType, content, optJSON, correctAnswer, points)
	return err
}

func (s *Store) DeleteQuizQuestion(ctx context.Context, id, quizID int) error {
	_, err := s.DB.Exec(ctx, `DELETE FROM quiz_question WHERE id=$1 AND quiz_id=$2`, id, quizID)
	return err
}

// ─── Quiz Attempts ──────────────────────────────────────

func (s *Store) GetStudentAttempt(ctx context.Context, quizID, studentID int) (*QuizAttempt, error) {
	a := &QuizAttempt{}
	var answersJSON, fileAnswersJSON []byte
	err := s.DB.QueryRow(ctx, `
		SELECT id, quiz_id, student_id, answers, file_answers, score, max_score, reviewed, started_at, finished_at
		FROM quiz_attempt WHERE quiz_id=$1 AND student_id=$2 ORDER BY started_at DESC LIMIT 1`,
		quizID, studentID).
		Scan(&a.ID, &a.QuizID, &a.StudentID, &answersJSON, &fileAnswersJSON, &a.Score, &a.MaxScore, &a.Reviewed, &a.StartedAt, &a.FinishedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	json.Unmarshal(answersJSON, &a.Answers)
	json.Unmarshal(fileAnswersJSON, &a.FileAnswers)
	return a, nil
}

func (s *Store) ListQuizAttempts(ctx context.Context, quizID int) ([]QuizAttempt, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT a.id, a.quiz_id, a.student_id, s.name, a.answers, a.file_answers, a.score, a.max_score, a.reviewed, a.started_at, a.finished_at
		FROM quiz_attempt a JOIN student s ON a.student_id=s.id
		WHERE a.quiz_id=$1 ORDER BY a.started_at DESC`, quizID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []QuizAttempt
	for rows.Next() {
		var a QuizAttempt
		var answersJSON, fileAnswersJSON []byte
		if err := rows.Scan(&a.ID, &a.QuizID, &a.StudentID, &a.StudentName, &answersJSON, &fileAnswersJSON,
			&a.Score, &a.MaxScore, &a.Reviewed, &a.StartedAt, &a.FinishedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(answersJSON, &a.Answers)
		json.Unmarshal(fileAnswersJSON, &a.FileAnswers)
		list = append(list, a)
	}
	return list, nil
}

func (s *Store) ListQuizAttemptsPaged(ctx context.Context, quizID, limit, offset int) ([]QuizAttempt, int, error) {
	var total int
	s.DB.QueryRow(ctx, `SELECT COUNT(*) FROM quiz_attempt WHERE quiz_id=$1`, quizID).Scan(&total)

	rows, err := s.DB.Query(ctx, `
		SELECT a.id, a.quiz_id, a.student_id, s.name, a.answers, a.file_answers, a.score, a.max_score, a.reviewed, a.started_at, a.finished_at
		FROM quiz_attempt a JOIN student s ON a.student_id=s.id
		WHERE a.quiz_id=$1 ORDER BY a.started_at DESC
		LIMIT $2 OFFSET $3`, quizID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []QuizAttempt
	for rows.Next() {
		var a QuizAttempt
		var answersJSON, fileAnswersJSON []byte
		if err := rows.Scan(&a.ID, &a.QuizID, &a.StudentID, &a.StudentName, &answersJSON, &fileAnswersJSON,
			&a.Score, &a.MaxScore, &a.Reviewed, &a.StartedAt, &a.FinishedAt); err != nil {
			return nil, 0, err
		}
		json.Unmarshal(answersJSON, &a.Answers)
		json.Unmarshal(fileAnswersJSON, &a.FileAnswers)
		list = append(list, a)
	}
	return list, total, nil
}

func (s *Store) CreateQuizAttempt(ctx context.Context, quizID, studentID int) (int, error) {
	var id int
	err := s.DB.QueryRow(ctx,
		`INSERT INTO quiz_attempt (quiz_id, student_id) VALUES ($1,$2) RETURNING id`,
		quizID, studentID).Scan(&id)
	return id, err
}

func (s *Store) SubmitQuizAttempt(ctx context.Context, id int, answers map[string]string, fileAnswers map[string]map[string]string, score, maxScore int) error {
	answersJSON, _ := json.Marshal(answers)
	fileAnswersJSON, _ := json.Marshal(fileAnswers)
	_, err := s.DB.Exec(ctx,
		`UPDATE quiz_attempt SET answers=$2, file_answers=$3, score=$4, max_score=$5, finished_at=NOW() WHERE id=$1`,
		id, answersJSON, fileAnswersJSON, score, maxScore)
	return err
}

func (s *Store) CountStudentAttempts(ctx context.Context, quizID, studentID int) (int, error) {
	var count int
	err := s.DB.QueryRow(ctx,
		`SELECT COUNT(*) FROM quiz_attempt WHERE quiz_id=$1 AND student_id=$2 AND finished_at IS NOT NULL`,
		quizID, studentID).Scan(&count)
	return count, err
}

func (s *Store) GetAllStudentAttempts(ctx context.Context, quizID, studentID int) ([]QuizAttempt, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, quiz_id, student_id, '', answers, file_answers, score, max_score, reviewed, started_at, finished_at
		FROM quiz_attempt WHERE quiz_id=$1 AND student_id=$2 AND finished_at IS NOT NULL ORDER BY started_at DESC`,
		quizID, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []QuizAttempt
	for rows.Next() {
		var a QuizAttempt
		var answersJSON, fileAnswersJSON []byte
		if err := rows.Scan(&a.ID, &a.QuizID, &a.StudentID, &a.StudentName, &answersJSON, &fileAnswersJSON,
			&a.Score, &a.MaxScore, &a.Reviewed, &a.StartedAt, &a.FinishedAt); err != nil {
			return nil, err
		}
		json.Unmarshal(answersJSON, &a.Answers)
		json.Unmarshal(fileAnswersJSON, &a.FileAnswers)
		list = append(list, a)
	}
	return list, nil
}

func (s *Store) ReviewQuizAttempt(ctx context.Context, id, score, classroomID int) error {
	_, err := s.DB.Exec(ctx, `UPDATE quiz_attempt SET score=$2, reviewed=true
		WHERE id=$1 AND quiz_id IN (SELECT id FROM quiz WHERE classroom_id=$3)`,
		id, score, classroomID)
	return err
}

// ─── Live Sessions ──────────────────────────────────────

func (s *Store) GetActiveLiveSession(ctx context.Context, classroomID int) (*LiveSession, error) {
	ls := &LiveSession{}
	err := s.DB.QueryRow(ctx,
		`SELECT id, classroom_id, room_name, active, ended_at, duration_minutes, created_at FROM live_session WHERE classroom_id=$1 AND active=true ORDER BY created_at DESC LIMIT 1`,
		classroomID).Scan(&ls.ID, &ls.ClassroomID, &ls.RoomName, &ls.Active, &ls.EndedAt, &ls.DurationMinutes, &ls.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return ls, nil
}

func (s *Store) CreateLiveSession(ctx context.Context, classroomID int, roomName string) (int, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	// End any existing active sessions first
	tx.Exec(ctx, `UPDATE live_session SET active=false, ended_at=NOW(),
		duration_minutes = EXTRACT(EPOCH FROM (NOW() - created_at))::int / 60
		WHERE classroom_id=$1 AND active=true`, classroomID)
	// Keep old inactive sessions for attendance history (no delete)
	var id int
	err = tx.QueryRow(ctx,
		`INSERT INTO live_session (classroom_id, room_name) VALUES ($1, $2) RETURNING id`,
		classroomID, roomName).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, tx.Commit(ctx)
}

func (s *Store) EndLiveSession(ctx context.Context, classroomID int) error {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Mark all attendance records as left
	tx.Exec(ctx, `
		UPDATE live_attendance SET left_at = NOW()
		WHERE left_at IS NULL
		AND live_session_id IN (SELECT id FROM live_session WHERE classroom_id=$1 AND active=true)`,
		classroomID)
	// End the session with timestamp and duration
	_, err = tx.Exec(ctx, `
		UPDATE live_session
		SET active=false, ended_at=NOW(),
		    duration_minutes = EXTRACT(EPOCH FROM (NOW() - created_at))::int / 60
		WHERE classroom_id=$1 AND active=true`, classroomID)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// ─── Live Attendance ────────────────────────────────────

type LiveAttendance struct {
	ID            int
	LiveSessionID int
	StudentID     int
	StudentName   string
	JoinedAt      time.Time
	LeftAt        *time.Time
	DurationSecs  int
}

type SessionHistoryRow struct {
	ID              int
	ClassroomID     int
	CreatedAt       time.Time
	EndedAt         *time.Time
	DurationMinutes *int
	AttendeeCount   int
}

type StudentAttendanceRow struct {
	StudentID      int
	Name           string
	Email          string
	SessionsJoined int
	TotalSessions  int
	AttendancePct  float64
	TotalMinutes   int
}

func (s *Store) JoinLiveAttendance(ctx context.Context, sessionID, studentID int) (int, error) {
	// Check if already has an open attendance record
	var existingID int
	err := s.DB.QueryRow(ctx,
		`SELECT id FROM live_attendance WHERE live_session_id=$1 AND student_id=$2 AND left_at IS NULL`,
		sessionID, studentID).Scan(&existingID)
	if err == nil {
		return existingID, nil // already joined
	}
	var id int
	err = s.DB.QueryRow(ctx,
		`INSERT INTO live_attendance (live_session_id, student_id) VALUES ($1, $2) RETURNING id`,
		sessionID, studentID).Scan(&id)
	return id, err
}

func (s *Store) LeaveLiveAttendance(ctx context.Context, sessionID, studentID int) error {
	_, err := s.DB.Exec(ctx,
		`UPDATE live_attendance SET left_at = NOW() WHERE live_session_id=$1 AND student_id=$2 AND left_at IS NULL`,
		sessionID, studentID)
	return err
}

func (s *Store) GetSessionHistory(ctx context.Context, classroomID int) ([]SessionHistoryRow, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT ls.id, ls.classroom_id, ls.created_at, ls.ended_at, ls.duration_minutes,
			(SELECT COUNT(DISTINCT student_id) FROM live_attendance WHERE live_session_id=ls.id) AS attendee_count
		FROM live_session ls
		WHERE ls.classroom_id = $1
		ORDER BY ls.created_at DESC`, classroomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []SessionHistoryRow
	for rows.Next() {
		var r SessionHistoryRow
		if err := rows.Scan(&r.ID, &r.ClassroomID, &r.CreatedAt, &r.EndedAt, &r.DurationMinutes, &r.AttendeeCount); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, nil
}

func (s *Store) GetSessionAttendance(ctx context.Context, sessionID int) ([]LiveAttendance, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT la.id, la.live_session_id, la.student_id, st.name, la.joined_at, la.left_at,
			COALESCE(EXTRACT(EPOCH FROM (COALESCE(la.left_at, NOW()) - la.joined_at))::int, 0) AS duration_secs
		FROM live_attendance la
		JOIN student st ON la.student_id = st.id
		WHERE la.live_session_id = $1
		ORDER BY la.joined_at`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []LiveAttendance
	for rows.Next() {
		var a LiveAttendance
		if err := rows.Scan(&a.ID, &a.LiveSessionID, &a.StudentID, &a.StudentName,
			&a.JoinedAt, &a.LeftAt, &a.DurationSecs); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, nil
}

func (s *Store) GetStudentAttendanceRates(ctx context.Context, classroomID int) ([]StudentAttendanceRow, error) {
	// Total sessions for this classroom
	var totalSessions int
	s.DB.QueryRow(ctx,
		`SELECT COUNT(*) FROM live_session WHERE classroom_id=$1`, classroomID).Scan(&totalSessions)

	rows, err := s.DB.Query(ctx, `
		SELECT s.id, s.name, COALESCE(s.email, ''),
			COUNT(DISTINCT la.live_session_id) AS sessions_joined,
			COALESCE(SUM(EXTRACT(EPOCH FROM (COALESCE(la.left_at, NOW()) - la.joined_at))::int / 60), 0) AS total_minutes
		FROM student s
		JOIN classroom_student cs ON s.id = cs.student_id
		LEFT JOIN live_attendance la ON la.student_id = s.id
			AND la.live_session_id IN (SELECT id FROM live_session WHERE classroom_id = $1)
		WHERE cs.classroom_id = $1 AND cs.status = 'approved'
		GROUP BY s.id, s.name, s.email
		ORDER BY sessions_joined DESC, s.name`, classroomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []StudentAttendanceRow
	for rows.Next() {
		var r StudentAttendanceRow
		if err := rows.Scan(&r.StudentID, &r.Name, &r.Email, &r.SessionsJoined, &r.TotalMinutes); err != nil {
			return nil, err
		}
		r.TotalSessions = totalSessions
		if totalSessions > 0 {
			r.AttendancePct = float64(r.SessionsJoined) * 100.0 / float64(totalSessions)
		}
		list = append(list, r)
	}
	return list, nil
}

// ─── Analytics Models ───────────────────────────────────

type QuizAnalyticsSummary struct {
	QuizID       int
	Title        string
	AttemptCount int
	StudentCount int
	AvgPct       float64
	HighestPct   float64
	LowestPct    float64
}

type QuestionAnalytics struct {
	QuestionID   int
	Content      string
	QuestionType string
	Points       int
	CorrectCount int
	TotalCount   int
	CorrectPct   float64
	CommonWrong  string
}

type AssignmentAnalyticsSummary struct {
	AssignmentID    int
	Title           string
	MaxGrade        float64
	Deadline        *time.Time
	SubmissionCount int
	EnrolledCount   int
	GradedCount     int
	AvgGrade        float64
	AvgPct          float64
	GradeBins       [5]int // A(90-100%), B(80-89%), C(70-79%), D(60-69%), F(<60%)
}

type MissingStudent struct {
	ID    int
	Name  string
	Email string
}

type StudentRosterRow struct {
	StudentID            int
	Name                 string
	Email                string
	AvgQuizPct           float64
	QuizzesTaken         int
	QuizzesTotal         int
	AssignmentsSubmitted int
	AssignmentsTotal     int
	AvgAssignmentPct     float64
	EngagementPct        float64
}

type StudentQuizDetail struct {
	QuizID    int
	QuizTitle string
	Score     *int
	MaxScore  *int
	Pct       float64
	StartedAt time.Time
	Duration  string
}

type StudentAssignmentDetail struct {
	AssignmentID int
	Title        string
	Grade        *float64
	MaxGrade     float64
	Pct          float64
	Status       string
	SubmittedAt  time.Time
}

// ─── Analytics Queries ──────────────────────────────────

func (s *Store) GetQuizAnalytics(ctx context.Context, classroomID int) ([]QuizAnalyticsSummary, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT q.id, q.title,
			COUNT(DISTINCT a.id) AS attempt_count,
			COUNT(DISTINCT a.student_id) AS student_count,
			COALESCE(AVG(CASE WHEN a.max_score > 0 THEN a.score * 100.0 / a.max_score ELSE 0 END), 0) AS avg_pct,
			COALESCE(MAX(CASE WHEN a.max_score > 0 THEN a.score * 100.0 / a.max_score ELSE 0 END), 0) AS highest_pct,
			COALESCE(MIN(CASE WHEN a.max_score > 0 THEN a.score * 100.0 / a.max_score ELSE 0 END), 0) AS lowest_pct
		FROM quiz q
		LEFT JOIN quiz_attempt a ON a.quiz_id = q.id AND a.finished_at IS NOT NULL
		WHERE q.classroom_id = $1
		GROUP BY q.id, q.title, q.created_at
		ORDER BY q.created_at DESC`, classroomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []QuizAnalyticsSummary
	for rows.Next() {
		var s QuizAnalyticsSummary
		if err := rows.Scan(&s.QuizID, &s.Title, &s.AttemptCount, &s.StudentCount,
			&s.AvgPct, &s.HighestPct, &s.LowestPct); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

func (s *Store) GetQuestionAnalytics(ctx context.Context, quizID int) ([]QuestionAnalytics, error) {
	// Get questions
	questions, err := s.ListQuizQuestions(ctx, quizID)
	if err != nil {
		return nil, err
	}
	// Get all finished attempts
	attempts, err := s.ListQuizAttempts(ctx, quizID)
	if err != nil {
		return nil, err
	}

	var results []QuestionAnalytics
	for _, q := range questions {
		qa := QuestionAnalytics{
			QuestionID:   q.ID,
			Content:      q.Content,
			QuestionType: q.QuestionType,
			Points:       q.Points,
		}

		if q.QuestionType == "open_ended" || q.QuestionType == "file_upload" {
			qa.TotalCount = len(attempts)
			results = append(results, qa)
			continue
		}

		wrongCounts := map[string]int{}
		for _, att := range attempts {
			if att.FinishedAt == nil {
				continue
			}
			qa.TotalCount++
			ans := att.Answers[fmt.Sprintf("%d", q.ID)]
			if strings.EqualFold(strings.TrimSpace(ans), strings.TrimSpace(q.CorrectAnswer)) {
				qa.CorrectCount++
			} else if ans != "" {
				wrongCounts[ans]++
			}
		}

		if qa.TotalCount > 0 {
			qa.CorrectPct = float64(qa.CorrectCount) * 100.0 / float64(qa.TotalCount)
		}

		// Find most common wrong answer
		maxWrong := 0
		for ans, cnt := range wrongCounts {
			if cnt > maxWrong {
				maxWrong = cnt
				qa.CommonWrong = ans
			}
		}

		results = append(results, qa)
	}
	return results, nil
}

func (s *Store) GetQuizStudentBreakdown(ctx context.Context, quizID int) ([]QuizAttempt, error) {
	return s.ListQuizAttempts(ctx, quizID)
}

func (s *Store) GetAssignmentAnalytics(ctx context.Context, classroomID int) ([]AssignmentAnalyticsSummary, error) {
	// Get enrolled count
	var enrolledCount int
	s.DB.QueryRow(ctx,
		`SELECT COUNT(*) FROM classroom_student WHERE classroom_id=$1 AND status='approved'`,
		classroomID).Scan(&enrolledCount)

	rows, err := s.DB.Query(ctx, `
		SELECT a.id, a.title, a.max_grade, a.deadline,
			COUNT(DISTINCT sub.id) AS submission_count,
			COUNT(DISTINCT CASE WHEN sub.grade IS NOT NULL THEN sub.id END) AS graded_count,
			COALESCE(AVG(sub.grade) FILTER (WHERE sub.grade IS NOT NULL), 0) AS avg_grade,
			COUNT(*) FILTER (WHERE sub.grade IS NOT NULL AND sub.max_grade > 0 AND sub.grade * 100.0 / sub.max_grade >= 90) AS bin_a,
			COUNT(*) FILTER (WHERE sub.grade IS NOT NULL AND sub.max_grade > 0 AND sub.grade * 100.0 / sub.max_grade >= 80 AND sub.grade * 100.0 / sub.max_grade < 90) AS bin_b,
			COUNT(*) FILTER (WHERE sub.grade IS NOT NULL AND sub.max_grade > 0 AND sub.grade * 100.0 / sub.max_grade >= 70 AND sub.grade * 100.0 / sub.max_grade < 80) AS bin_c,
			COUNT(*) FILTER (WHERE sub.grade IS NOT NULL AND sub.max_grade > 0 AND sub.grade * 100.0 / sub.max_grade >= 60 AND sub.grade * 100.0 / sub.max_grade < 70) AS bin_d,
			COUNT(*) FILTER (WHERE sub.grade IS NOT NULL AND sub.max_grade > 0 AND sub.grade * 100.0 / sub.max_grade < 60) AS bin_f
		FROM assignment a
		LEFT JOIN submission sub ON sub.assignment_id = a.id
		WHERE a.classroom_id = $1
		GROUP BY a.id, a.title, a.max_grade, a.deadline, a.created_at
		ORDER BY a.created_at DESC`, classroomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []AssignmentAnalyticsSummary
	for rows.Next() {
		var a AssignmentAnalyticsSummary
		if err := rows.Scan(&a.AssignmentID, &a.Title, &a.MaxGrade, &a.Deadline,
			&a.SubmissionCount, &a.GradedCount, &a.AvgGrade,
			&a.GradeBins[0], &a.GradeBins[1], &a.GradeBins[2], &a.GradeBins[3], &a.GradeBins[4]); err != nil {
			return nil, err
		}
		a.EnrolledCount = enrolledCount
		if a.MaxGrade > 0 && a.GradedCount > 0 {
			a.AvgPct = a.AvgGrade * 100.0 / a.MaxGrade
		}
		list = append(list, a)
	}

	return list, nil
}

func (s *Store) GetMissingSubmissions(ctx context.Context, assignmentID, classroomID int) ([]MissingStudent, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT s.id, s.name, COALESCE(s.email, '')
		FROM student s
		JOIN classroom_student cs ON s.id = cs.student_id
		WHERE cs.classroom_id = $1 AND cs.status = 'approved'
		AND s.id NOT IN (SELECT student_id FROM submission WHERE assignment_id = $2)
		ORDER BY s.name`, classroomID, assignmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []MissingStudent
	for rows.Next() {
		var m MissingStudent
		if err := rows.Scan(&m.ID, &m.Name, &m.Email); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, nil
}

func (s *Store) GetStudentRosterAnalytics(ctx context.Context, classroomID int) ([]StudentRosterRow, error) {
	rows, err := s.DB.Query(ctx, `
		WITH totals AS (
			SELECT
				(SELECT COUNT(*) FROM quiz WHERE classroom_id=$1 AND published=true) AS quizzes_total,
				(SELECT COUNT(*) FROM assignment WHERE classroom_id=$1) AS assignments_total
		),
		students AS (
			SELECT s.id, s.name, COALESCE(s.email, '') AS email
			FROM student s
			JOIN classroom_student cs ON s.id = cs.student_id
			WHERE cs.classroom_id = $1 AND cs.status = 'approved'
		),
		quiz_stats AS (
			SELECT student_id,
				COUNT(DISTINCT quiz_id) AS quizzes_taken,
				COALESCE(AVG(best_pct), 0) AS avg_quiz_pct
			FROM (
				SELECT qa.student_id, qa.quiz_id,
					MAX(CASE WHEN qa.max_score > 0 THEN qa.score * 100.0 / qa.max_score ELSE 0 END) AS best_pct
				FROM quiz_attempt qa
				JOIN quiz q ON qa.quiz_id = q.id
				WHERE q.classroom_id = $1 AND qa.finished_at IS NOT NULL
				GROUP BY qa.student_id, qa.quiz_id
			) sub
			GROUP BY student_id
		),
		assign_graded AS (
			SELECT sub.student_id,
				COUNT(DISTINCT sub.assignment_id) AS graded_count,
				COALESCE(AVG(CASE WHEN sub.max_grade > 0 THEN sub.grade * 100.0 / sub.max_grade ELSE 0 END), 0) AS avg_assign_pct
			FROM submission sub
			JOIN assignment a ON sub.assignment_id = a.id
			WHERE a.classroom_id = $1 AND sub.grade IS NOT NULL
			GROUP BY sub.student_id
		),
		assign_submitted AS (
			SELECT sub.student_id,
				COUNT(DISTINCT sub.assignment_id) AS submitted_count
			FROM submission sub
			JOIN assignment a ON sub.assignment_id = a.id
			WHERE a.classroom_id = $1
			GROUP BY sub.student_id
		)
		SELECT st.id, st.name, st.email,
			t.quizzes_total, t.assignments_total,
			COALESCE(qs.quizzes_taken, 0),
			COALESCE(qs.avg_quiz_pct, 0),
			COALESCE(ag.graded_count, COALESCE(asub.submitted_count, 0)),
			COALESCE(ag.avg_assign_pct, 0)
		FROM students st
		CROSS JOIN totals t
		LEFT JOIN quiz_stats qs ON qs.student_id = st.id
		LEFT JOIN assign_graded ag ON ag.student_id = st.id
		LEFT JOIN assign_submitted asub ON asub.student_id = st.id
		ORDER BY st.name`, classroomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []StudentRosterRow
	for rows.Next() {
		var r StudentRosterRow
		if err := rows.Scan(&r.StudentID, &r.Name, &r.Email,
			&r.QuizzesTotal, &r.AssignmentsTotal,
			&r.QuizzesTaken, &r.AvgQuizPct,
			&r.AssignmentsSubmitted, &r.AvgAssignmentPct); err != nil {
			return nil, err
		}
		totalItems := r.QuizzesTotal + r.AssignmentsTotal
		doneItems := r.QuizzesTaken + r.AssignmentsSubmitted
		if totalItems > 0 {
			r.EngagementPct = float64(doneItems) * 100.0 / float64(totalItems)
		}
		list = append(list, r)
	}
	return list, nil
}

func (s *Store) GetStudentQuizDetails(ctx context.Context, studentID, classroomID int) ([]StudentQuizDetail, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT q.id, q.title, a.score, a.max_score, a.started_at, a.finished_at
		FROM quiz_attempt a
		JOIN quiz q ON a.quiz_id = q.id
		WHERE a.student_id = $1 AND q.classroom_id = $2 AND a.finished_at IS NOT NULL
		ORDER BY a.started_at DESC`, studentID, classroomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []StudentQuizDetail
	for rows.Next() {
		var d StudentQuizDetail
		var finishedAt *time.Time
		if err := rows.Scan(&d.QuizID, &d.QuizTitle, &d.Score, &d.MaxScore, &d.StartedAt, &finishedAt); err != nil {
			return nil, err
		}
		if d.MaxScore != nil && *d.MaxScore > 0 && d.Score != nil {
			d.Pct = float64(*d.Score) * 100.0 / float64(*d.MaxScore)
		}
		if finishedAt != nil {
			dur := finishedAt.Sub(d.StartedAt)
			mins := int(dur.Minutes())
			secs := int(dur.Seconds()) % 60
			if mins > 0 {
				d.Duration = fmt.Sprintf("%dm %ds", mins, secs)
			} else {
				d.Duration = fmt.Sprintf("%ds", secs)
			}
		}
		list = append(list, d)
	}
	return list, nil
}

func (s *Store) GetStudentAssignmentDetails(ctx context.Context, studentID, classroomID int) ([]StudentAssignmentDetail, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT a.id, a.title, sub.grade, a.max_grade, sub.status, sub.submitted_at
		FROM submission sub
		JOIN assignment a ON sub.assignment_id = a.id
		WHERE sub.student_id = $1 AND a.classroom_id = $2
		ORDER BY sub.submitted_at DESC`, studentID, classroomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []StudentAssignmentDetail
	for rows.Next() {
		var d StudentAssignmentDetail
		if err := rows.Scan(&d.AssignmentID, &d.Title, &d.Grade, &d.MaxGrade, &d.Status, &d.SubmittedAt); err != nil {
			return nil, err
		}
		if d.Grade != nil && d.MaxGrade > 0 {
			d.Pct = *d.Grade * 100.0 / d.MaxGrade
		}
		list = append(list, d)
	}
	return list, nil
}

// ─── Student Remarks ────────────────────────────────────

type StudentRemark struct {
	ID          int
	ClassroomID int
	StudentID   int
	Content     string
	CreatedAt   time.Time
}

func (s *Store) GetStudentRemarks(ctx context.Context, studentID, classroomID int) ([]StudentRemark, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, classroom_id, student_id, content, created_at
		FROM student_remark
		WHERE student_id = $1 AND classroom_id = $2
		ORDER BY created_at DESC`, studentID, classroomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []StudentRemark
	for rows.Next() {
		var r StudentRemark
		if err := rows.Scan(&r.ID, &r.ClassroomID, &r.StudentID, &r.Content, &r.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, nil
}

func (s *Store) AddStudentRemark(ctx context.Context, classroomID, studentID int, content string) error {
	_, err := s.DB.Exec(ctx,
		`INSERT INTO student_remark (classroom_id, student_id, content) VALUES ($1, $2, $3)`,
		classroomID, studentID, content)
	return err
}

func (s *Store) DeleteStudentRemark(ctx context.Context, remarkID, classroomID int) error {
	_, err := s.DB.Exec(ctx, `DELETE FROM student_remark WHERE id=$1 AND classroom_id=$2`, remarkID, classroomID)
	return err
}

// ─── Student Dashboard Data ─────────────────────────────

type StudentAttendanceRecord struct {
	SessionID    int
	SessionDate  time.Time
	Duration     *int // session duration in minutes
	Attended     bool
	JoinedAt     *time.Time
	LeftAt       *time.Time
	TimeSpentMin int
}

func (s *Store) GetStudentAttendanceRecord(ctx context.Context, studentID, classroomID int) ([]StudentAttendanceRecord, int, int, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT ls.id, ls.created_at, ls.duration_minutes,
			la.id IS NOT NULL AS attended,
			la.joined_at, la.left_at,
			COALESCE(EXTRACT(EPOCH FROM (COALESCE(la.left_at, ls.ended_at, NOW()) - la.joined_at))::int / 60, 0) AS time_spent_min
		FROM live_session ls
		LEFT JOIN live_attendance la ON la.live_session_id = ls.id AND la.student_id = $1
		WHERE ls.classroom_id = $2
		  AND ls.active = false
		  AND COALESCE(ls.duration_minutes, 0) >= 5
		ORDER BY ls.created_at DESC`, studentID, classroomID)
	if err != nil {
		return nil, 0, 0, err
	}
	defer rows.Close()
	var list []StudentAttendanceRecord
	attended := 0
	total := 0
	for rows.Next() {
		var r StudentAttendanceRecord
		if err := rows.Scan(&r.SessionID, &r.SessionDate, &r.Duration, &r.Attended, &r.JoinedAt, &r.LeftAt, &r.TimeSpentMin); err != nil {
			return nil, 0, 0, err
		}
		total++
		if r.Attended {
			attended++
		}
		list = append(list, r)
	}
	return list, attended, total, nil
}

// GetStudentDashboardStats computes class averages so the student can see above/below indicator
type StudentDashboardStats struct {
	ClassAvgQuizPct   float64
	ClassAvgAssignPct float64
}

func (s *Store) GetStudentDashboardStats(ctx context.Context, classroomID int) (*StudentDashboardStats, error) {
	stats := &StudentDashboardStats{}
	// Class average quiz score %
	s.DB.QueryRow(ctx, `
		SELECT COALESCE(AVG(pct), 0) FROM (
			SELECT MAX(CASE WHEN a.max_score > 0 THEN a.score * 100.0 / a.max_score ELSE 0 END) AS pct
			FROM quiz_attempt a
			JOIN quiz q ON a.quiz_id = q.id
			WHERE q.classroom_id = $1 AND a.finished_at IS NOT NULL
			GROUP BY a.student_id, a.quiz_id
		) sub`, classroomID).Scan(&stats.ClassAvgQuizPct)

	// Class average assignment grade %
	s.DB.QueryRow(ctx, `
		SELECT COALESCE(AVG(CASE WHEN sub.max_grade > 0 THEN sub.grade * 100.0 / sub.max_grade ELSE 0 END), 0)
		FROM submission sub
		JOIN assignment a ON sub.assignment_id = a.id
		WHERE a.classroom_id = $1 AND sub.grade IS NOT NULL`,
		classroomID).Scan(&stats.ClassAvgAssignPct)

	return stats, nil
}

// ─── Resource View Tracking ─────────────────────────────

func (s *Store) TrackResourceView(ctx context.Context, resourceID, studentID int) error {
	_, err := s.DB.Exec(ctx,
		`INSERT INTO resource_view (resource_id, student_id)
		 SELECT $1, $2 WHERE NOT EXISTS (
		   SELECT 1 FROM resource_view WHERE resource_id=$1 AND student_id=$2 AND viewed_at > NOW() - INTERVAL '1 hour'
		 )`,
		resourceID, studentID)
	return err
}

type ResourceViewCount struct {
	ResourceID     int
	Title          string
	FileType       string
	ViewCount      int
	UniqueStudents int
}

func (s *Store) GetResourceViewCounts(ctx context.Context, classroomID int) ([]ResourceViewCount, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT r.id, r.title, COALESCE(r.file_type, 'file'),
			COUNT(rv.id) AS view_count,
			COUNT(DISTINCT rv.student_id) AS unique_students
		FROM resource r
		LEFT JOIN resource_view rv ON rv.resource_id = r.id
		WHERE r.classroom_id = $1
		GROUP BY r.id, r.title, r.file_type, r.created_at
		ORDER BY view_count DESC, r.created_at DESC`, classroomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []ResourceViewCount
	for rows.Next() {
		var r ResourceViewCount
		if err := rows.Scan(&r.ResourceID, &r.Title, &r.FileType, &r.ViewCount, &r.UniqueStudents); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, nil
}

// ─── Performance Trends ─────────────────────────────────

type QuizTrendPoint struct {
	QuizID    int
	Title     string
	AvgPct    float64
	CreatedAt time.Time
}

func (s *Store) GetQuizTrends(ctx context.Context, classroomID int) ([]QuizTrendPoint, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT q.id, q.title,
			COALESCE(AVG(CASE WHEN a.max_score > 0 THEN a.score * 100.0 / a.max_score ELSE 0 END), 0) AS avg_pct,
			q.created_at
		FROM quiz q
		LEFT JOIN quiz_attempt a ON a.quiz_id = q.id AND a.finished_at IS NOT NULL
		WHERE q.classroom_id = $1
		GROUP BY q.id, q.title, q.created_at
		ORDER BY q.created_at ASC`, classroomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []QuizTrendPoint
	for rows.Next() {
		var p QuizTrendPoint
		if err := rows.Scan(&p.QuizID, &p.Title, &p.AvgPct, &p.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

type AssignmentTrendPoint struct {
	AssignmentID int
	Title        string
	AvgPct       float64
	CreatedAt    time.Time
}

func (s *Store) GetAssignmentTrends(ctx context.Context, classroomID int) ([]AssignmentTrendPoint, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT a.id, a.title,
			COALESCE(AVG(CASE WHEN a.max_grade > 0 THEN sub.grade * 100.0 / a.max_grade ELSE 0 END), 0) AS avg_pct,
			a.created_at
		FROM assignment a
		LEFT JOIN submission sub ON sub.assignment_id = a.id AND sub.grade IS NOT NULL
		WHERE a.classroom_id = $1
		GROUP BY a.id, a.title, a.created_at
		ORDER BY a.created_at ASC`, classroomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []AssignmentTrendPoint
	for rows.Next() {
		var p AssignmentTrendPoint
		if err := rows.Scan(&p.AssignmentID, &p.Title, &p.AvgPct, &p.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

// ─── At-Risk Students ───────────────────────────────────

type AtRiskStudent struct {
	StudentID int
	Name      string
	Email     string
	Reasons   []string
	RiskScore int // higher = more at risk
}

func (s *Store) GetAtRiskStudents(ctx context.Context, classroomID int) ([]AtRiskStudent, error) {
	rows, err := s.DB.Query(ctx, `
		WITH totals AS (
			SELECT
				(SELECT COUNT(*) FROM assignment WHERE classroom_id=$1) AS total_assignments,
				(SELECT COUNT(*) FROM live_session WHERE classroom_id=$1) AS total_sessions
		),
		students AS (
			SELECT s.id, s.name, COALESCE(s.email, '') AS email
			FROM student s
			JOIN classroom_student cs ON s.id = cs.student_id
			WHERE cs.classroom_id = $1 AND cs.status = 'approved'
		),
		submission_counts AS (
			SELECT sub.student_id, COUNT(DISTINCT sub.assignment_id) AS submitted
			FROM submission sub
			JOIN assignment a ON sub.assignment_id = a.id
			WHERE a.classroom_id = $1
			GROUP BY sub.student_id
		),
		recent_quiz_scores AS (
			SELECT student_id,
				COUNT(*) FILTER (WHERE pct < 50) AS low_count
			FROM (
				SELECT qa.student_id,
					CASE WHEN qa.max_score > 0 THEN qa.score * 100.0 / qa.max_score ELSE 0 END AS pct,
					ROW_NUMBER() OVER (PARTITION BY qa.student_id ORDER BY qa.finished_at DESC) AS rn
				FROM quiz_attempt qa
				JOIN quiz q ON qa.quiz_id = q.id
				WHERE q.classroom_id = $1 AND qa.finished_at IS NOT NULL
			) ranked WHERE rn <= 2
			GROUP BY student_id
		),
		attendance_counts AS (
			SELECT la.student_id, COUNT(DISTINCT la.live_session_id) AS attended
			FROM live_attendance la
			JOIN live_session ls ON la.live_session_id = ls.id
			WHERE ls.classroom_id = $1
			GROUP BY la.student_id
		)
		SELECT st.id, st.name, st.email,
			t.total_assignments, COALESCE(sc.submitted, 0) AS submitted,
			COALESCE(rqs.low_count, 0) AS low_quiz_count,
			t.total_sessions, COALESCE(ac.attended, 0) AS attended
		FROM students st
		CROSS JOIN totals t
		LEFT JOIN submission_counts sc ON sc.student_id = st.id
		LEFT JOIN recent_quiz_scores rqs ON rqs.student_id = st.id
		LEFT JOIN attendance_counts ac ON ac.student_id = st.id`, classroomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var atRisk []AtRiskStudent
	for rows.Next() {
		var (
			id, totalAssignments, submitted, lowQuizCount, totalSessions, attended int
			name, email                                                            string
		)
		if err := rows.Scan(&id, &name, &email, &totalAssignments, &submitted, &lowQuizCount, &totalSessions, &attended); err != nil {
			return nil, err
		}

		ar := AtRiskStudent{StudentID: id, Name: name, Email: email}

		// Missed 2+ assignments
		missed := totalAssignments - submitted
		if missed >= 2 {
			ar.Reasons = append(ar.Reasons, fmt.Sprintf("Missed %d assignments", missed))
			ar.RiskScore += 2
		}

		// Scored <50% on last 2 quizzes
		if lowQuizCount >= 2 {
			ar.Reasons = append(ar.Reasons, "Scored <50% on last 2 quizzes")
			ar.RiskScore += 3
		}

		// Attendance <50%
		if totalSessions > 0 {
			pct := float64(attended) * 100.0 / float64(totalSessions)
			if pct < 50.0 {
				ar.Reasons = append(ar.Reasons, fmt.Sprintf("Attendance %.0f%%", pct))
				ar.RiskScore += 2
			}
		}

		if len(ar.Reasons) > 0 {
			atRisk = append(atRisk, ar)
		}
	}

	sort.Slice(atRisk, func(i, j int) bool {
		return atRisk[i].RiskScore > atRisk[j].RiskScore
	})

	return atRisk, nil
}

// ─── Submission Timing ──────────────────────────────────

type SubmissionTimingStats struct {
	AssignmentID int
	Title        string
	EarlyCount   int // submitted >24h before deadline
	OnTimeCount  int // submitted <24h before deadline
	LateCount    int // submitted after deadline (if any)
	NoDeadline   bool
}

func (s *Store) GetSubmissionTimingStats(ctx context.Context, classroomID int) ([]SubmissionTimingStats, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT a.id, a.title,
			a.deadline IS NULL AS no_deadline,
			COUNT(*) FILTER (WHERE sub.id IS NOT NULL AND a.deadline IS NOT NULL AND sub.submitted_at <= a.deadline - INTERVAL '24 hours') AS early_count,
			COUNT(*) FILTER (WHERE sub.id IS NOT NULL AND (
				(a.deadline IS NOT NULL AND sub.submitted_at > a.deadline - INTERVAL '24 hours' AND sub.submitted_at <= a.deadline)
				OR a.deadline IS NULL
			)) AS ontime_count,
			COUNT(*) FILTER (WHERE sub.id IS NOT NULL AND a.deadline IS NOT NULL AND sub.submitted_at > a.deadline) AS late_count
		FROM assignment a
		LEFT JOIN submission sub ON sub.assignment_id = a.id
		WHERE a.classroom_id = $1
		GROUP BY a.id, a.title, a.deadline, a.created_at
		ORDER BY a.created_at DESC`, classroomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []SubmissionTimingStats
	for rows.Next() {
		var st SubmissionTimingStats
		if err := rows.Scan(&st.AssignmentID, &st.Title, &st.NoDeadline,
			&st.EarlyCount, &st.OnTimeCount, &st.LateCount); err != nil {
			return nil, err
		}
		list = append(list, st)
	}
	return list, nil
}

// ─── CSV Export Queries ─────────────────────────────────

type RosterExportRow struct {
	Name                 string
	Email                string
	AvgQuizPct           float64
	QuizzesTaken         int
	AvgAssignmentPct     float64
	AssignmentsSubmitted int
	AttendancePct        float64
	EngagementPct        float64
}

func (s *Store) GetRosterExport(ctx context.Context, classroomID int) ([]RosterExportRow, error) {
	roster, err := s.GetStudentRosterAnalytics(ctx, classroomID)
	if err != nil {
		return nil, err
	}
	// Get attendance rates
	attRates, _ := s.GetStudentAttendanceRates(ctx, classroomID)
	attMap := map[int]float64{}
	for _, a := range attRates {
		attMap[a.StudentID] = a.AttendancePct
	}

	var list []RosterExportRow
	for _, r := range roster {
		list = append(list, RosterExportRow{
			Name:                 r.Name,
			Email:                r.Email,
			AvgQuizPct:           r.AvgQuizPct,
			QuizzesTaken:         r.QuizzesTaken,
			AvgAssignmentPct:     r.AvgAssignmentPct,
			AssignmentsSubmitted: r.AssignmentsSubmitted,
			AttendancePct:        attMap[r.StudentID],
			EngagementPct:        r.EngagementPct,
		})
	}
	return list, nil
}

type QuizExportRow struct {
	QuizTitle    string
	StudentName  string
	StudentEmail string
	Score        int
	MaxScore     int
	Pct          float64
	StartedAt    time.Time
	FinishedAt   *time.Time
}

func (s *Store) GetQuizExport(ctx context.Context, classroomID int) ([]QuizExportRow, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT q.title, s.name, COALESCE(s.email, ''),
			COALESCE(a.score, 0), COALESCE(a.max_score, 0),
			CASE WHEN a.max_score > 0 THEN a.score * 100.0 / a.max_score ELSE 0 END AS pct,
			a.started_at, a.finished_at
		FROM quiz_attempt a
		JOIN quiz q ON a.quiz_id = q.id
		JOIN student s ON a.student_id = s.id
		WHERE q.classroom_id = $1 AND a.finished_at IS NOT NULL
		ORDER BY q.title, s.name, a.started_at`, classroomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []QuizExportRow
	for rows.Next() {
		var r QuizExportRow
		if err := rows.Scan(&r.QuizTitle, &r.StudentName, &r.StudentEmail, &r.Score, &r.MaxScore, &r.Pct, &r.StartedAt, &r.FinishedAt); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, nil
}

type AssignmentExportRow struct {
	AssignmentTitle string
	StudentName     string
	StudentEmail    string
	Grade           *float64
	MaxGrade        float64
	Pct             float64
	Status          string
	SubmittedAt     time.Time
}

func (s *Store) GetAssignmentExport(ctx context.Context, classroomID int) ([]AssignmentExportRow, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT a.title, s.name, COALESCE(s.email, ''),
			sub.grade, a.max_grade,
			CASE WHEN a.max_grade > 0 AND sub.grade IS NOT NULL THEN sub.grade * 100.0 / a.max_grade ELSE 0 END AS pct,
			sub.status, sub.submitted_at
		FROM submission sub
		JOIN assignment a ON sub.assignment_id = a.id
		JOIN student s ON sub.student_id = s.id
		WHERE a.classroom_id = $1
		ORDER BY a.title, s.name`, classroomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []AssignmentExportRow
	for rows.Next() {
		var r AssignmentExportRow
		if err := rows.Scan(&r.AssignmentTitle, &r.StudentName, &r.StudentEmail, &r.Grade, &r.MaxGrade, &r.Pct, &r.Status, &r.SubmittedAt); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, nil
}

type AttendanceExportRow struct {
	StudentName     string
	StudentEmail    string
	SessionDate     time.Time
	SessionDuration *int
	JoinedAt        *time.Time
	LeftAt          *time.Time
	TimeSpentMin    int
	Attended        bool
}

func (s *Store) GetAttendanceExport(ctx context.Context, classroomID int) ([]AttendanceExportRow, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT s.name, COALESCE(s.email, ''),
			ls.created_at, ls.duration_minutes,
			la.joined_at, la.left_at,
			COALESCE(EXTRACT(EPOCH FROM (COALESCE(la.left_at, ls.ended_at, NOW()) - la.joined_at))::int / 60, 0) AS time_spent_min,
			la.id IS NOT NULL AS attended
		FROM live_session ls
		CROSS JOIN (
			SELECT DISTINCT s2.id, s2.name, s2.email
			FROM student s2
			JOIN classroom_student cs ON s2.id = cs.student_id
			WHERE cs.classroom_id = $1 AND cs.status = 'approved'
		) s
		LEFT JOIN live_attendance la ON la.live_session_id = ls.id AND la.student_id = s.id
		WHERE ls.classroom_id = $1
		ORDER BY ls.created_at DESC, s.name`, classroomID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []AttendanceExportRow
	for rows.Next() {
		var r AttendanceExportRow
		if err := rows.Scan(&r.StudentName, &r.StudentEmail, &r.SessionDate, &r.SessionDuration, &r.JoinedAt, &r.LeftAt, &r.TimeSpentMin, &r.Attended); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, nil
}

// ─── Platform Admin ─────────────────────────────────────

// ─── Platform Analytics ─────────────────────────────────

type TopTeacher struct {
	ID                 int
	Username           string
	Email              string
	SchoolName         string
	SubscriptionStatus string
	ClassroomCount     int
	StudentCount       int
	QuizCount          int
	ResourceCount      int
}

type MonthlyRevenueRow struct {
	Month   string // "2026-03"
	Revenue float64
}

type ApplicationTrendRow struct {
	Month string
	Count int
}

func (s *Store) TotalStudentsOnPlatform(ctx context.Context) (int, error) {
	var count int
	err := s.DB.QueryRow(ctx, `SELECT COUNT(DISTINCT id) FROM student`).Scan(&count)
	return count, err
}

func (s *Store) TotalClassroomsOnPlatform(ctx context.Context) (int, error) {
	var count int
	err := s.DB.QueryRow(ctx, `SELECT COUNT(*) FROM classroom`).Scan(&count)
	return count, err
}

func (s *Store) TotalQuizzesOnPlatform(ctx context.Context) (int, error) {
	var count int
	err := s.DB.QueryRow(ctx, `SELECT COUNT(*) FROM quiz`).Scan(&count)
	return count, err
}

func (s *Store) TopTeachersByStudents(ctx context.Context, limit int) ([]TopTeacher, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT a.id, a.username, a.email, a.school_name, a.subscription_status,
			(SELECT COUNT(*) FROM classroom WHERE admin_id=a.id) AS classroom_count,
			(SELECT COUNT(DISTINCT cs.student_id) FROM classroom_student cs JOIN classroom c2 ON c2.id=cs.classroom_id WHERE c2.admin_id=a.id AND cs.status='approved') AS student_count,
			(SELECT COUNT(*) FROM quiz q JOIN classroom c3 ON c3.id=q.classroom_id WHERE c3.admin_id=a.id) AS quiz_count,
			(SELECT COUNT(*) FROM resource r JOIN classroom c4 ON c4.id=r.classroom_id WHERE c4.admin_id=a.id) AS resource_count
		FROM admin a
		WHERE a.created_by_platform = true
		ORDER BY student_count DESC, a.subscription_start DESC NULLS LAST
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []TopTeacher
	for rows.Next() {
		var t TopTeacher
		if err := rows.Scan(&t.ID, &t.Username, &t.Email, &t.SchoolName, &t.SubscriptionStatus,
			&t.ClassroomCount, &t.StudentCount, &t.QuizCount, &t.ResourceCount); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, nil
}

func (s *Store) MonthlyRevenueBreakdown(ctx context.Context, months int) ([]MonthlyRevenueRow, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT to_char(m.month, 'YYYY-MM') AS month_label,
		       COALESCE(SUM(p.amount), 0) AS revenue
		FROM generate_series(
			date_trunc('month', NOW()) - ($1 - 1) * interval '1 month',
			date_trunc('month', NOW()),
			interval '1 month'
		) AS m(month)
		LEFT JOIN payment p ON date_trunc('month', p.recorded_at) = m.month
		GROUP BY m.month
		ORDER BY m.month`, months)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []MonthlyRevenueRow
	for rows.Next() {
		var r MonthlyRevenueRow
		if err := rows.Scan(&r.Month, &r.Revenue); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, nil
}

func (s *Store) ApplicationsTrend(ctx context.Context, months int) ([]ApplicationTrendRow, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT to_char(m.month, 'YYYY-MM') AS month_label,
		       COUNT(ta.id) AS cnt
		FROM generate_series(
			date_trunc('month', NOW()) - ($1 - 1) * interval '1 month',
			date_trunc('month', NOW()),
			interval '1 month'
		) AS m(month)
		LEFT JOIN teacher_application ta ON date_trunc('month', ta.created_at) = m.month
		GROUP BY m.month
		ORDER BY m.month`, months)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []ApplicationTrendRow
	for rows.Next() {
		var r ApplicationTrendRow
		if err := rows.Scan(&r.Month, &r.Count); err != nil {
			return nil, err
		}
		list = append(list, r)
	}
	return list, nil
}

func (s *Store) ListAllPayments(ctx context.Context) ([]Payment, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT p.id, p.teacher_id, p.amount, p.method, p.reference, p.notes, p.recorded_at
		FROM payment p ORDER BY p.recorded_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Payment
	for rows.Next() {
		var p Payment
		if err := rows.Scan(&p.ID, &p.TeacherID, &p.Amount, &p.Method, &p.Reference, &p.Notes, &p.RecordedAt); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func (s *Store) UpdatePlatformAdminPassword(ctx context.Context, id int, hashedPassword string) error {
	_, err := s.DB.Exec(ctx,
		`UPDATE platform_admin SET password_hash=$1 WHERE id=$2`,
		hashedPassword, id)
	return err
}

func (s *Store) GetPlatformAdminByID(ctx context.Context, id int) (*PlatformAdmin, error) {
	a := &PlatformAdmin{}
	err := s.DB.QueryRow(ctx,
		`SELECT id, username, password_hash, created_at FROM platform_admin WHERE id=$1`, id).
		Scan(&a.ID, &a.Username, &a.PasswordHash, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Store) GetAdminByID(ctx context.Context, id int) (*Admin, error) {
	a := &Admin{}
	err := s.DB.QueryRow(ctx,
		`SELECT id, username, password, email, school_name, subscription_status,
		        subscription_start, subscription_end, created_by_platform, application_id, pending_password
		 FROM admin WHERE id=$1`, id).
		Scan(&a.ID, &a.Username, &a.Password, &a.Email, &a.SchoolName,
			&a.SubscriptionStatus, &a.SubscriptionStart, &a.SubscriptionEnd,
			&a.CreatedByPlatform, &a.ApplicationID, &a.PendingPassword)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Store) CreateTeacherFromApplication(ctx context.Context, username, hashedPassword, plaintextPassword, email, schoolName string, applicationID int) (int, error) {
	var id int
	err := s.DB.QueryRow(ctx,
		`INSERT INTO admin (username, password, pending_password, email, school_name, subscription_status, subscription_start, created_by_platform, application_id)
		 VALUES ($1, $2, $3, $4, $5, 'active', NOW(), true, $6) RETURNING id`,
		username, hashedPassword, plaintextPassword, email, schoolName, applicationID).Scan(&id)
	return id, err
}

func (s *Store) ClearPendingPassword(ctx context.Context, adminID int) {
	s.DB.Exec(ctx, `UPDATE admin SET pending_password = NULL WHERE id = $1`, adminID)
}

func (s *Store) ListTeachers(ctx context.Context) ([]TeacherListItem, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT a.id, a.username, a.email, a.school_name, a.subscription_status,
		       a.subscription_start, a.subscription_end,
		       COALESCE((SELECT COUNT(*) FROM classroom WHERE admin_id = a.id), 0) AS classroom_count,
		       COALESCE((SELECT COUNT(DISTINCT cs.student_id) FROM classroom_student cs JOIN classroom c ON c.id = cs.classroom_id WHERE c.admin_id = a.id AND cs.status='approved'), 0) AS student_count
		FROM admin a
		WHERE a.created_by_platform = true
		ORDER BY a.subscription_start DESC NULLS LAST`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []TeacherListItem
	for rows.Next() {
		var t TeacherListItem
		if err := rows.Scan(&t.ID, &t.Username, &t.Email, &t.SchoolName, &t.SubscriptionStatus,
			&t.SubscriptionStart, &t.SubscriptionEnd, &t.ClassroomCount, &t.StudentCount); err != nil {
			return nil, err
		}
		list = append(list, t)
	}
	return list, nil
}

func (s *Store) GetTeacherStats(ctx context.Context, teacherID int) (classrooms, students, quizzes, resources int, err error) {
	err = s.DB.QueryRow(ctx, `SELECT COUNT(*) FROM classroom WHERE admin_id=$1`, teacherID).Scan(&classrooms)
	if err != nil {
		return
	}
	err = s.DB.QueryRow(ctx, `SELECT COUNT(DISTINCT cs.student_id) FROM classroom_student cs JOIN classroom c ON c.id=cs.classroom_id WHERE c.admin_id=$1 AND cs.status='approved'`, teacherID).Scan(&students)
	if err != nil {
		return
	}
	err = s.DB.QueryRow(ctx, `SELECT COUNT(*) FROM quiz q JOIN classroom c ON c.id=q.classroom_id WHERE c.admin_id=$1`, teacherID).Scan(&quizzes)
	if err != nil {
		return
	}
	err = s.DB.QueryRow(ctx, `SELECT COUNT(*) FROM resource r JOIN classroom c ON c.id=r.classroom_id WHERE c.admin_id=$1`, teacherID).Scan(&resources)
	return
}

func (s *Store) UpdateTeacherSubscription(ctx context.Context, teacherID int, status string) error {
	_, err := s.DB.Exec(ctx,
		`UPDATE admin SET subscription_status=$1 WHERE id=$2`,
		status, teacherID)
	return err
}

func (s *Store) SetSubscriptionEnd(ctx context.Context, teacherID int, endDate time.Time) error {
	_, err := s.DB.Exec(ctx,
		`UPDATE admin SET subscription_end=$1 WHERE id=$2`,
		endDate, teacherID)
	return err
}

func (s *Store) ExtendSubscription(ctx context.Context, teacherID int, months int) error {
	_, err := s.DB.Exec(ctx,
		`UPDATE admin SET
			subscription_status = 'active',
			subscription_end = CASE
				WHEN subscription_end IS NULL OR subscription_end < NOW() THEN NOW() + ($1 || ' months')::interval
				ELSE subscription_end + ($1 || ' months')::interval
			END
		 WHERE id = $2`,
		months, teacherID)
	return err
}

// ─── Payments ───────────────────────────────────────────

func (s *Store) CreatePayment(ctx context.Context, teacherID int, amount float64, method, reference, notes string) error {
	_, err := s.DB.Exec(ctx,
		`INSERT INTO payment (teacher_id, amount, method, reference, notes) VALUES ($1, $2, $3, $4, $5)`,
		teacherID, amount, method, reference, notes)
	return err
}

func (s *Store) ListPaymentsByTeacher(ctx context.Context, teacherID int) ([]Payment, error) {
	rows, err := s.DB.Query(ctx,
		`SELECT id, teacher_id, amount, method, reference, notes, recorded_at
		 FROM payment WHERE teacher_id=$1 ORDER BY recorded_at DESC`, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Payment
	for rows.Next() {
		var p Payment
		if err := rows.Scan(&p.ID, &p.TeacherID, &p.Amount, &p.Method, &p.Reference, &p.Notes, &p.RecordedAt); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func (s *Store) DeletePayment(ctx context.Context, paymentID int) error {
	_, err := s.DB.Exec(ctx, `DELETE FROM payment WHERE id=$1`, paymentID)
	return err
}

func (s *Store) GetTeacherTotalPayments(ctx context.Context, teacherID int) (float64, int, error) {
	var total float64
	var count int
	err := s.DB.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount), 0), COUNT(*) FROM payment WHERE teacher_id=$1`, teacherID).
		Scan(&total, &count)
	return total, count, err
}

func (s *Store) CountActiveTeachers(ctx context.Context) (active, suspended, expired int, err error) {
	rows, err := s.DB.Query(ctx,
		`SELECT subscription_status, COUNT(*) FROM admin WHERE created_by_platform=true GROUP BY subscription_status`)
	if err != nil {
		return 0, 0, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var st string
		var cnt int
		if err := rows.Scan(&st, &cnt); err != nil {
			return 0, 0, 0, err
		}
		switch st {
		case "active":
			active = cnt
		case "suspended":
			suspended = cnt
		case "expired":
			expired = cnt
		}
	}
	return
}

func (s *Store) CountExpiringSoon(ctx context.Context, withinDays int) (int, error) {
	var count int
	err := s.DB.QueryRow(ctx,
		`SELECT COUNT(*) FROM admin
		 WHERE created_by_platform=true
		   AND subscription_status='active'
		   AND subscription_end IS NOT NULL
		   AND subscription_end <= NOW() + ($1 || ' days')::interval
		   AND subscription_end > NOW()`, withinDays).Scan(&count)
	return count, err
}

func (s *Store) GetTotalRevenue(ctx context.Context) (float64, error) {
	var total float64
	err := s.DB.QueryRow(ctx, `SELECT COALESCE(SUM(amount), 0) FROM payment`).Scan(&total)
	return total, err
}

func (s *Store) GetMonthlyRevenue(ctx context.Context) (float64, error) {
	var total float64
	err := s.DB.QueryRow(ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM payment
		 WHERE recorded_at >= date_trunc('month', NOW())`).Scan(&total)
	return total, err
}

func (s *Store) CheckAndExpireSubscriptions(ctx context.Context) (int, error) {
	tag, err := s.DB.Exec(ctx,
		`UPDATE admin SET subscription_status='expired'
		 WHERE created_by_platform=true
		   AND subscription_status='active'
		   AND subscription_end IS NOT NULL
		   AND subscription_end < NOW()`)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

func (s *Store) GetPlatformAdminByUsername(ctx context.Context, username string) (*PlatformAdmin, error) {
	a := &PlatformAdmin{}
	err := s.DB.QueryRow(ctx,
		`SELECT id, username, password_hash, created_at FROM platform_admin WHERE username=$1`, username).
		Scan(&a.ID, &a.Username, &a.PasswordHash, &a.CreatedAt)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Store) CreatePlatformAdmin(ctx context.Context, username, hashedPassword string) error {
	_, err := s.DB.Exec(ctx,
		`INSERT INTO platform_admin (username, password_hash) VALUES ($1, $2) ON CONFLICT (username) DO NOTHING`,
		username, hashedPassword)
	return err
}

// ─── Teacher Applications ───────────────────────────────

func (s *Store) CreateTeacherApplication(ctx context.Context, fullName, email, phone, school, wilaya, message string) error {
	_, err := s.DB.Exec(ctx,
		`INSERT INTO teacher_application (full_name, email, phone, school_name, wilaya, message)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		fullName, email, phone, school, wilaya, message)
	return err
}

func (s *Store) ListTeacherApplications(ctx context.Context, statusFilter string) ([]TeacherApplication, error) {
	query := `SELECT id, full_name, email, phone, school_name, wilaya, message, status, admin_notes, created_at, reviewed_at
	          FROM teacher_application`
	var args []interface{}
	if statusFilter != "" && statusFilter != "all" {
		query += ` WHERE status = $1`
		args = append(args, statusFilter)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := s.DB.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []TeacherApplication
	for rows.Next() {
		var a TeacherApplication
		if err := rows.Scan(&a.ID, &a.FullName, &a.Email, &a.Phone, &a.SchoolName, &a.Wilaya, &a.Message, &a.Status, &a.AdminNotes, &a.CreatedAt, &a.ReviewedAt); err != nil {
			return nil, err
		}
		list = append(list, a)
	}
	return list, nil
}

func (s *Store) GetTeacherApplication(ctx context.Context, id int) (*TeacherApplication, error) {
	a := &TeacherApplication{}
	err := s.DB.QueryRow(ctx,
		`SELECT id, full_name, email, phone, school_name, wilaya, message, status, admin_notes, created_at, reviewed_at
		 FROM teacher_application WHERE id=$1`, id).
		Scan(&a.ID, &a.FullName, &a.Email, &a.Phone, &a.SchoolName, &a.Wilaya, &a.Message, &a.Status, &a.AdminNotes, &a.CreatedAt, &a.ReviewedAt)
	if err != nil {
		return nil, err
	}
	return a, nil
}

func (s *Store) UpdateApplicationStatus(ctx context.Context, id int, status, notes string) error {
	_, err := s.DB.Exec(ctx,
		`UPDATE teacher_application SET status=$1, admin_notes=$2, reviewed_at=NOW() WHERE id=$3`,
		status, notes, id)
	return err
}

func (s *Store) CountApplicationsByStatus(ctx context.Context) (pending, approved, rejected, contacted int, err error) {
	rows, err := s.DB.Query(ctx,
		`SELECT status, COUNT(*) FROM teacher_application GROUP BY status`)
	if err != nil {
		return 0, 0, 0, 0, err
	}
	defer rows.Close()
	for rows.Next() {
		var st string
		var cnt int
		if err := rows.Scan(&st, &cnt); err != nil {
			return 0, 0, 0, 0, err
		}
		switch st {
		case "pending":
			pending = cnt
		case "approved":
			approved = cnt
		case "rejected":
			rejected = cnt
		case "contacted":
			contacted = cnt
		}
	}
	return
}
