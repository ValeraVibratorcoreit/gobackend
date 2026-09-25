package main

import (
	"encoding/json"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

type App struct {
	store *Store
}

func NewApp(store *Store) *App {
	return &App{store: store}
}

// ---------- вспомогательные функции ----------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v != nil {
		if err := json.NewEncoder(w).Encode(v); err != nil {
			log.Printf("write json: %v", err)
		}
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// внутренние ошибки БД никогда не уходят клиенту — только в лог
func writeInternalError(w http.ResponseWriter, err error) {
	log.Printf("internal error: %v", err)
	writeError(w, http.StatusInternalServerError, "internal server error")
}

func parseID(r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, false
	}
	return id, true
}

var timeRe = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)

var validStatuses = map[string]bool{
	"scheduled": true,
	"cancelled": true,
	"moved":     true,
}

// ---------- /lessons ----------

func (a *App) listLessons(w http.ResponseWriter, r *http.Request) {
	var groupID *int64
	if raw := r.URL.Query().Get("group_id"); raw != "" {
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || id <= 0 {
			writeError(w, http.StatusBadRequest, "group_id must be a positive integer")
			return
		}
		groupID = &id
	}

	lessons, err := a.store.ListLessons(groupID)
	if err != nil {
		writeInternalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, lessons)
}

func (a *App) getLesson(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "id must be a positive integer")
		return
	}

	lesson, err := a.store.GetLesson(id)
	if err == ErrNotFound {
		writeError(w, http.StatusNotFound, "lesson not found")
		return
	}
	if err != nil {
		writeInternalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, lesson)
}

func (a *App) createLesson(w http.ResponseWriter, r *http.Request) {
	var req CreateLessonRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if msg, ok := validateCreateLesson(req); !ok {
		writeError(w, http.StatusUnprocessableEntity, msg)
		return
	}

	lesson, err := a.store.CreateLesson(req)
	if err == ErrConstraint {
		writeError(w, http.StatusUnprocessableEntity, "invalid subject_id, teacher_id, group_id or classroom_id")
		return
	}
	if err != nil {
		writeInternalError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, lesson)
}

func (a *App) updateLessonStatus(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "id must be a positive integer")
		return
	}

	var req UpdateLessonStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if !validStatuses[req.Status] {
		writeError(w, http.StatusUnprocessableEntity, "status must be one of: scheduled, cancelled, moved")
		return
	}

	lesson, err := a.store.UpdateLessonStatus(id, req.Status)
	if err == ErrNotFound {
		writeError(w, http.StatusNotFound, "lesson not found")
		return
	}
	if err == ErrConstraint {
		writeError(w, http.StatusUnprocessableEntity, "invalid status")
		return
	}
	if err != nil {
		writeInternalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, lesson)
}

func (a *App) deleteLesson(w http.ResponseWriter, r *http.Request) {
	id, ok := parseID(r)
	if !ok {
		writeError(w, http.StatusBadRequest, "id must be a positive integer")
		return
	}

	err := a.store.DeleteLesson(id)
	if err == ErrNotFound {
		writeError(w, http.StatusNotFound, "lesson not found")
		return
	}
	if err != nil {
		writeInternalError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func validateCreateLesson(req CreateLessonRequest) (string, bool) {
	if req.SubjectID <= 0 {
		return "subject_id is required", false
	}
	if req.TeacherID <= 0 {
		return "teacher_id is required", false
	}
	if req.GroupID <= 0 {
		return "group_id is required", false
	}
	if req.ClassroomID != nil && *req.ClassroomID <= 0 {
		return "classroom_id must be a positive integer", false
	}
	if req.DayOfWeek < 1 || req.DayOfWeek > 7 {
		return "day_of_week must be between 1 and 7", false
	}
	if !timeRe.MatchString(req.StartTime) {
		return "start_time must be in HH:MM format", false
	}
	if !timeRe.MatchString(req.EndTime) {
		return "end_time must be in HH:MM format", false
	}
	if req.EndTime <= req.StartTime {
		return "end_time must be after start_time", false
	}
	return "", true
}

// ---------- справочники (только чтение) ----------

func (a *App) listGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := a.store.ListGroups()
	if err != nil {
		writeInternalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, groups)
}

func (a *App) listTeachers(w http.ResponseWriter, r *http.Request) {
	teachers, err := a.store.ListTeachers()
	if err != nil {
		writeInternalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, teachers)
}

func (a *App) listSubjects(w http.ResponseWriter, r *http.Request) {
	subjects, err := a.store.ListSubjects()
	if err != nil {
		writeInternalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, subjects)
}

func (a *App) listClassrooms(w http.ResponseWriter, r *http.Request) {
	classrooms, err := a.store.ListClassrooms()
	if err != nil {
		writeInternalError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, classrooms)
}

// ---------- health ----------

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
