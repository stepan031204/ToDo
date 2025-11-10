// ---- DOM элементы ----
const tasksContainer = document.getElementById('tasksContainer');
const addForm = document.getElementById('addForm');
const nameInput = document.getElementById('nameInput');
const statusEl = document.getElementById('status');

// ---- Вспомогательные функции ----

function showStatus(text, timeout = 3000) {
  statusEl.textContent = text;
  if (timeout > 0) {
    setTimeout(() => {
      if (statusEl.textContent === text) statusEl.textContent = '';
    }, timeout);
  }
}

async function apiFetch(url, options = {}) {
  const res = await fetch(url, options);
  if (!res.ok) {
    const text = await res.text().catch(() => res.statusText);
    throw new Error(`HTTP ${res.status}: ${text}`);
  }
  const ct = res.headers.get('Content-Type') || '';
  if (ct.includes('application/json')) {
    return await res.json();
  }
  return null;
}

// ---- Основная логика ----

async function loadTasks() {
  try {
    const data = await apiFetch('/api/tasks');
    renderTasks(data || []);
  } catch (err) {
    showStatus('Ошибка загрузки задач: ' + err.message);
    console.error(err);
  }
}

function createTaskElement(task) {
  const wrap = document.createElement('div');
  wrap.className = 'task';

  const left = document.createElement('div');
  left.className = 'left';

  const checkbox = document.createElement('div');
  checkbox.className = 'checkbox';
  checkbox.title = task.done ? 'Отменить выполнение' : 'Отметить как выполненное';
  checkbox.textContent = task.done ? '✔' : '';

  const name = document.createElement('div');
  name.className = 'name' + (task.done ? ' done' : '');
  name.textContent = `${task.id}. ${task.name}`;

  left.appendChild(checkbox);
  left.appendChild(name);

  const actions = document.createElement('div');
  actions.className = 'actions';

  const del = document.createElement('a');
  del.textContent = '🗑';
  del.href = '#';
  del.title = 'Удалить задачу';

  actions.appendChild(del);
  wrap.appendChild(left);
  wrap.appendChild(actions);

  // --- Обработчики ---

  checkbox.classList.toggle('checked', task.done);

// при клике на чекбокс
checkbox.addEventListener('click', async (e) => {
  e.preventDefault();
  task.done = !task.done;

  // плавно обновляем UI
  checkbox.classList.toggle('checked', task.done);
  name.classList.toggle('done', task.done);

  try {
    await apiFetch('/api/toggle', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id: task.id })
    });
    showStatus('Статус обновлён');
  } catch (err) {
    // если ошибка, откатываем UI
    task.done = !task.done;
    checkbox.classList.toggle('checked', task.done);
    name.classList.toggle('done', task.done);
    showStatus('Не удалось обновить статус: ' + err.message);
    console.error(err);
  }
});
  del.addEventListener('click', async (e) => {
    e.preventDefault();
    if (!confirm('Удалить задачу?')) return;
    try {
      await apiFetch(`/api/delete?id=${task.id}`, { method: 'DELETE' });
      wrap.remove();
      showStatus('Задача удалена');
    } catch (err) {
      showStatus('Ошибка удаления: ' + err.message);
    }
  });

  return wrap;
}

function renderTasks(tasks) {
  tasksContainer.innerHTML = '';
  if (!tasks.length) {
    const p = document.createElement('p');
    p.textContent = 'Пока нет задач ✨';
    tasksContainer.appendChild(p);
    return;
  }
  for (const t of tasks) {
    const el = createTaskElement(t);
    tasksContainer.appendChild(el);
  }
}

// ---- Обработка добавления ----

addForm.addEventListener('submit', async (e) => {
  e.preventDefault();
  const name = nameInput.value.trim();
  if (!name) return;

  try {
    await apiFetch('/api/add', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name })
    });
    nameInput.value = '';
    showStatus('Задача добавлена');
    await loadTasks();
  } catch (err) {
    showStatus('Ошибка при добавлении: ' + err.message);
  }
});

// ---- Запуск ----
loadTasks();
