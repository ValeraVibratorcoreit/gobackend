package main

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

// ErrNotFound — запись не найдена.
var ErrNotFound = errors.New("not found")

// ErrConstraint — нарушено ограничение БД (FK или CHECK):
// несуществующий subject/teacher/group/classroom, либо day_of_week/время вне допустимых значений.
var ErrConstraint = errors.New("constraint violation")

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

const lessonSelectBase = `
	SELECT
		l.id, l.subject_id, s.name,
		l.teacher_id, t.full_name,
		l.group_id, g.name,
		l.classroom_id, c.name,
		l.day_of_week,
		to_char(l.start_time, 'HH24:MI'),
		to_char(l.end_time, 'HH24:MI'),
		l.status, l.created_at
	FROM lessons l
	JOIN subjects s ON s.id = l.subject_id
	JOIN teachers t ON t.id = l.teacher_id
	JOIN groups g ON g.id = l.group_id
	LEFT JOIN classrooms c ON c.id = l.classroom_id
`

func scanLesson(row interface{ Scan(...any) error }) (*Lesson, error) {
	var l Lesson
	var classroomID sql.NullInt64
	var classroomName sql.NullString

	err := row.Scan(
		&l.ID, &l.SubjectID, &l.SubjectName,
		&l.TeacherID, &l.TeacherName,
		&l.GroupID, &l.GroupName,
		&classroomID, &classroomName,
		&l.DayOfWeek, &l.StartTime, &l.EndTime,
		&l.Status, &l.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if classroomID.Valid {
		v := classroomID.Int64
		l.ClassroomID = &v
	}
	if classroomName.Valid {
		v := classroomName.String
		l.ClassroomName = &v
	}
	return &l, nil
}

// ListLessons возвращает занятия, опционально отфильтрованные по группе.
func (s *Store) ListLessons(groupID *int64) ([]Lesson, error) {
	query := lessonSelectBase
	args := []any{}
	if groupID != nil {
		query += " WHERE l.group_id = $1"
		args = append(args, *groupID)
	}
	query += " ORDER BY l.day_of_week, l.start_time"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	lessons := []Lesson{}
	for rows.Next() {
		l, err := scanLesson(rows)
		if err != nil {
			return nil, err
		}
		lessons = append(lessons, *l)
	}
	return lessons, rows.Err()
}

func (s *Store) GetLesson(id int64) (*Lesson, error) {
	query := lessonSelectBase + " WHERE l.id = $1"
	row := s.db.QueryRow(query, id)
	l, err := scanLesson(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (s *Store) CreateLesson(req CreateLessonRequest) (*Lesson, error) {
	var id int64
	err := s.db.QueryRow(`
		INSERT INTO lessons (subject_id, teacher_id, group_id, classroom_id, day_of_week, start_time, end_time)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, req.SubjectID, req.TeacherID, req.GroupID, req.ClassroomID, req.DayOfWeek, req.StartTime, req.EndTime).Scan(&id)

	if err != nil {
		if isConstraintViolation(err) {
			return nil, ErrConstraint
		}
		return nil, err
	}
	return s.GetLesson(id)
}

func (s *Store) UpdateLessonStatus(id int64, status string) (*Lesson, error) {
	res, err := s.db.Exec(`UPDATE lessons SET status = $1 WHERE id = $2`, status, id)
	if err != nil {
		if isConstraintViolation(err) {
			return nil, ErrConstraint
		}
		return nil, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, ErrNotFound
	}
	return s.GetLesson(id)
}

func (s *Store) DeleteLesson(id int64) error {
	res, err := s.db.Exec(`DELETE FROM lessons WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListGroups() ([]Group, error) {
	rows, err := s.db.Query(`SELECT id, name FROM groups ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Group{}
	for rows.Next() {
		var g Group
		if err := rows.Scan(&g.ID, &g.Name); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

func (s *Store) ListTeachers() ([]Teacher, error) {
	rows, err := s.db.Query(`SELECT id, full_name, email FROM teachers ORDER BY full_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Teacher{}
	for rows.Next() {
		var t Teacher
		var email sql.NullString
		if err := rows.Scan(&t.ID, &t.FullName, &email); err != nil {
			return nil, err
		}
		if email.Valid {
			v := email.String
			t.Email = &v
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (s *Store) ListSubjects() ([]Subject, error) {
	rows, err := s.db.Query(`SELECT id, name FROM subjects ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Subject{}
	for rows.Next() {
		var sub Subject
		if err := rows.Scan(&sub.ID, &sub.Name); err != nil {
			return nil, err
		}
		out = append(out, sub)
	}
	return out, rows.Err()
}

func (s *Store) ListClassrooms() ([]Classroom, error) {
	rows, err := s.db.Query(`SELECT id, name FROM classrooms ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []Classroom{}
	for rows.Next() {
		var c Classroom
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

// isConstraintViolation определяет, что ошибка вызвана нарушением FK или CHECK —
// такие ошибки превращаются в 422, а не в 500, и не показывают деталей клиенту.
func isConstraintViolation(err error) bool {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		switch pqErr.Code.Name() {
		case "foreign_key_violation", "check_violation", "not_null_violation", "invalid_text_representation":
			return true
		}
	}
	return false
}
