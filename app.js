const API_BASE = "http://localhost:8080";

const DAYS = {
  1: "Пн", 2: "Вт", 3: "Ср", 4: "Чт", 5: "Пт", 6: "Сб", 7: "Вс",
};

const STATUS_LABELS = {
  scheduled: "Идёт",
  cancelled: "Отменено",
  moved: "Перенесено",
};

const el = {
  errorBanner: document.getElementById("error-banner"),
  filterGroup: document.getElementById("filter-group"),
  lessonForm: document.getElementById("lesson-form"),
  fSubject: document.getElementById("f-subject"),
  fTeacher: document.getElementById("f-teacher"),
  fGroup: document.getElementById("f-group"),
  fClassroom: document.getElementById("f-classroom"),
  fDay: document.getElementById("f-day"),
  fStart: document.getElementById("f-start"),
  fEnd: document.getElementById("f-end"),
  lessonsBody: document.getElementById("lessons-body"),
  emptyState: document.getElementById("empty-state"),
};

function showError(message) {
  el.errorBanner.textContent = message;
  el.errorBanner.hidden = false;
}

function clearError() {
  el.errorBanner.hidden = true;
  el.errorBanner.textContent = "";
}

// Единая точка вызова API: показывает ошибку сервера пользователю,
// а не глотает её молча.
async function api(path, options = {}) {
  let response;
  try {
    response = await fetch(`${API_BASE}${path}`, {
      headers: { "Content-Type": "application/json" },
      ...options,
    });
  } catch (err) {
    showError("Не удалось связаться с сервером. Проверьте, что API запущен.");
    throw err;
  }

  if (response.status === 204) {
    return null;
  }

  let body = null;
  try {
    body = await response.json();
  } catch {
    // тело может отсутствовать
  }

  if (!response.ok) {
    const message = body && body.error ? body.error : `Ошибка сервера (${response.status})`;
    showError(message);
    throw new Error(message);
  }

  return body;
}

function fillSelect(select, items, { placeholder } = {}) {
  select.innerHTML = "";
  if (placeholder !== undefined) {
    const opt = document.createElement("option");
    opt.value = "";
    opt.textContent = placeholder;
    select.appendChild(opt);
  }
  for (const item of items) {
    const opt = document.createElement("option");
    opt.value = item.id;
    opt.textContent = item.name || item.full_name;
    select.appendChild(opt);
  }
}

async function loadReferenceData() {
  const [groups, teachers, subjects, classrooms] = await Promise.all([
    api("/groups"),
    api("/teachers"),
    api("/subjects"),
    api("/classrooms"),
  ]);

  fillSelect(el.filterGroup, groups, { placeholder: "Все группы" });
  fillSelect(el.fGroup, groups);
  fillSelect(el.fTeacher, teachers);
  fillSelect(el.fSubject, subjects);
  fillSelect(el.fClassroom, classrooms, { placeholder: "—" });
}

function renderLessons(lessons) {
  el.lessonsBody.innerHTML = "";
  el.emptyState.hidden = lessons.length > 0;

  for (const lesson of lessons) {
    const tr = document.createElement("tr");

    tr.innerHTML = `
      <td>${DAYS[lesson.day_of_week] || lesson.day_of_week}</td>
      <td>${lesson.start_time}–${lesson.end_time}</td>
      <td>${escapeHtml(lesson.subject_name)}</td>
      <td>${escapeHtml(lesson.teacher_name)}</td>
      <td>${escapeHtml(lesson.group_name)}</td>
      <td>${lesson.classroom_name ? escapeHtml(lesson.classroom_name) : "—"}</td>
      <td class="status status-${lesson.status}">${STATUS_LABELS[lesson.status] || lesson.status}</td>
      <td class="row-actions"></td>
    `;

    const actions = tr.querySelector(".row-actions");

    if (lesson.status !== "cancelled") {
      const cancelBtn = document.createElement("button");
      cancelBtn.className = "secondary";
      cancelBtn.textContent = "Отменить";
      cancelBtn.addEventListener("click", () => updateStatus(lesson.id, "cancelled"));
      actions.appendChild(cancelBtn);
    }

    const deleteBtn = document.createElement("button");
    deleteBtn.className = "secondary";
    deleteBtn.textContent = "Удалить";
    deleteBtn.addEventListener("click", () => deleteLesson(lesson.id));
    actions.appendChild(deleteBtn);

    el.lessonsBody.appendChild(tr);
  }
}

function escapeHtml(str) {
  const div = document.createElement("div");
  div.textContent = str ?? "";
  return div.innerHTML;
}

async function loadLessons() {
  clearError();
  const groupId = el.filterGroup.value;
  const path = groupId ? `/lessons?group_id=${encodeURIComponent(groupId)}` : "/lessons";
  try {
    const lessons = await api(path);
    renderLessons(lessons);
  } catch {
    // ошибка уже показана пользователю в api()
  }
}

async function updateStatus(id, status) {
  clearError();
  try {
    await api(`/lessons/${id}`, {
      method: "PATCH",
      body: JSON.stringify({ status }),
    });
    await loadLessons();
  } catch {
    // ошибка уже показана
  }
}

async function deleteLesson(id) {
  clearError();
  if (!confirm("Удалить занятие из расписания?")) return;
  try {
    await api(`/lessons/${id}`, { method: "DELETE" });
    await loadLessons();
  } catch {
    // ошибка уже показана
  }
}

el.lessonForm.addEventListener("submit", async (e) => {
  e.preventDefault();
  clearError();

  const payload = {
    subject_id: Number(el.fSubject.value),
    teacher_id: Number(el.fTeacher.value),
    group_id: Number(el.fGroup.value),
    classroom_id: el.fClassroom.value ? Number(el.fClassroom.value) : null,
    day_of_week: Number(el.fDay.value),
    start_time: el.fStart.value,
    end_time: el.fEnd.value,
  };

  try {
    await api("/lessons", { method: "POST", body: JSON.stringify(payload) });
    el.lessonForm.reset();
    await loadLessons();
  } catch {
    // ошибка уже показана
  }
});

el.filterGroup.addEventListener("change", loadLessons);

(async function init() {
  try {
    await loadReferenceData();
    await loadLessons();
  } catch {
    // ошибка уже показана в api()
  }
})();
