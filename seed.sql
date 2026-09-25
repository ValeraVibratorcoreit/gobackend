-- Тестовые данные для «Расписания занятий»

INSERT INTO groups (name) VALUES
    ('ИВТ-21'),
    ('ПИ-22');

INSERT INTO teachers (full_name, email) VALUES
    ('Иванова Анна Сергеевна', 'ivanova@university.edu'),
    ('Петров Игорь Викторович', 'petrov@university.edu'),
    ('Сидорова Ольга Павловна', NULL);

INSERT INTO subjects (name) VALUES
    ('Базы данных'),
    ('Алгоритмы и структуры данных'),
    ('Веб-разработка');

INSERT INTO classrooms (name) VALUES
    ('312'),
    ('412'),
    ('Комп. класс 1');

-- id: groups (1=ИВТ-21, 2=ПИ-22), teachers (1=Иванова, 2=Петров, 3=Сидорова),
-- subjects (1=БД, 2=Алгоритмы, 3=Веб), classrooms (1=312, 2=412, 3=Комп. класс 1)

INSERT INTO lessons (subject_id, teacher_id, group_id, classroom_id, day_of_week, start_time, end_time, status) VALUES
    (1, 1, 1, 2, 1, '09:00', '10:30', 'scheduled'),
    (2, 2, 1, 1, 1, '10:45', '12:15', 'scheduled'),
    (3, 3, 1, 3, 2, '09:00', '10:30', 'scheduled'),
    (1, 1, 2, 2, 3, '13:00', '14:30', 'moved'),
    (2, 2, 2, 1, 4, '09:00', '10:30', 'cancelled');
