package main

import "time"

// Lesson — одно занятие в расписании.
type Lesson struct {
	ID            int64     `json:"id"`
	SubjectID     int64     `json:"subject_id"`
	SubjectName   string    `json:"subject_name,omitempty"`
	TeacherID     int64     `json:"teacher_id"`
	TeacherName   string    `json:"teacher_name,omitempty"`
	GroupID       int64     `json:"group_id"`
	GroupName     string    `json:"group_name,omitempty"`
	ClassroomID   *int64    `json:"classroom_id"`
	ClassroomName *string   `json:"classroom_name,omitempty"`
	DayOfWeek     int16     `json:"day_of_week"`
	StartTime     string    `json:"start_time"` // формат "HH:MM"
	EndTime       string    `json:"end_time"`   // формат "HH:MM"
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}

type Group struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Teacher struct {
	ID       int64   `json:"id"`
	FullName string  `json:"full_name"`
	Email    *string `json:"email,omitempty"`
}

type Subject struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type Classroom struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// CreateLessonRequest — тело POST /lessons
type CreateLessonRequest struct {
	SubjectID   int64  `json:"subject_id"`
	TeacherID   int64  `json:"teacher_id"`
	GroupID     int64  `json:"group_id"`
	ClassroomID *int64 `json:"classroom_id"`
	DayOfWeek   int16  `json:"day_of_week"`
	StartTime   string `json:"start_time"`
	EndTime     string `json:"end_time"`
}

// UpdateLessonStatusRequest — тело PATCH /lessons/{id}
type UpdateLessonStatusRequest struct {
	Status string `json:"status"`
}
