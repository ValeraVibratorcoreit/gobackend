-- Схема БД: «Расписание занятий»
-- PostgreSQL

CREATE TABLE groups (
    id         SERIAL PRIMARY KEY,
    name       VARCHAR(50) NOT NULL UNIQUE,       -- например, "ИВТ-21"
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE teachers (
    id         SERIAL PRIMARY KEY,
    full_name  VARCHAR(150) NOT NULL,
    email      VARCHAR(150) UNIQUE,                -- может быть NULL, но если есть — уникален
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE subjects (
    id   SERIAL PRIMARY KEY,
    name VARCHAR(150) NOT NULL UNIQUE
);

CREATE TABLE classrooms (
    id   SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE               -- например, "412"
);

-- Основная сущность: конкретное занятие в расписании.
-- day_of_week: 1 = понедельник ... 7 = воскресенье
-- status: состояние занятия — то, что требуется по заданию
CREATE TABLE lessons (
    id           SERIAL PRIMARY KEY,
    subject_id   INT NOT NULL REFERENCES subjects(id) ON DELETE RESTRICT,
    teacher_id   INT NOT NULL REFERENCES teachers(id) ON DELETE RESTRICT,
    group_id     INT NOT NULL REFERENCES groups(id) ON DELETE RESTRICT,
    classroom_id INT REFERENCES classrooms(id) ON DELETE SET NULL,  -- аудитория может быть не назначена
    day_of_week  SMALLINT NOT NULL CHECK (day_of_week BETWEEN 1 AND 7),
    start_time   TIME NOT NULL,
    end_time     TIME NOT NULL CHECK (end_time > start_time),
    status       VARCHAR(20) NOT NULL DEFAULT 'scheduled'
                 CHECK (status IN ('scheduled', 'cancelled', 'moved')),
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Индексы на внешние ключи (обязательное требование критериев)
CREATE INDEX idx_lessons_subject_id   ON lessons(subject_id);
CREATE INDEX idx_lessons_teacher_id   ON lessons(teacher_id);
CREATE INDEX idx_lessons_group_id     ON lessons(group_id);
CREATE INDEX idx_lessons_classroom_id ON lessons(classroom_id);
