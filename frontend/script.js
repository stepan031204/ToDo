// ---- Константы DOM ----
const tasksContainer = document.getElementById('tasksContainer');
const addForm = document.getElementById('addForm');
const nameInput = document.getElementById('nameInput');
const statusEl = document.getElementById('status');

// ---- Вспомогательные функции ----

// helper: показывает текст состояния пользователю (и автоматически скрывает через 3s)
function showStatus(text, timeout = 3000) {
  statusEl.textContent = text;
  if (timeout > 0) {
    setTimeout(() => {
      // если статус не изменили за это время, очистим
      if (statusEl.textContent === text) statusEl.textContent = '';
    }, timeout);
  }
}

// helper: обёртка для fetch с проверкой статуса
async function apiFetch(url, options = {}) {
  const res = await fetch(url, options);           // отправляем запрос
  if (!res.ok) {
    const text = await res.text().catch(() => res.statusText);
    throw new Error(`HTTP ${res.status}: ${text}`);
  }
  // some endpoints return empty body (DELETE, POST add -> 201 with empty body), handle gracefully
  const ct = res.headers.get('Content-Type') || '';
  if (ct.includes('application/json')) {
    return await res.json();
  }
  return null;
}

// Загружает задачи с бэкенда и рендерит их в DOM
async function loadTasks() {
  try {
    const data = await apiFetch('/api/tasks');
    renderTasks(data || []);
  } catch (err) {
    showStatus('Ошибка загрузки задач: ' + err.message);
    console.error(err);
  }
}

// Создаёт DOM-элемент карточки задачи и навешивает обработчики
function createTaskElement(task) {
  // контейнер карточки
  const wrap = document.createElement('div');
  wrap.className = 'task';

  // левая часть: чекбокс + название
  const left = document.createElement('div');
  left.className = 'left';

  const checkbox = document.createElement('div');
  checkbox.className = 'checkbox';
  checkbox.title = task.Done ? 'Отменить выполнение' : 'Отметить как выполненное';
  checkbox.textContent = task.Done ? '✔' : '';

  const name = document.createElement('div');
  name.className = 'name' + (task.Done ? ' done' : '');
  name.textContent = `${task.ID}. ${task.Name}`;

  left.appendChild(checkbox);
  left.appendChild(name);

  // правая часть: кнопки действий
  const actions = document.createElement('div');
  actions.className = 'actions';

  const del = document.createElement('a');
  del.textContent = '🗑';
  del.href = '#';
  del.title = 'Удалить задачу';

  actions.appendChild(del);

  wrap.appendChild(left);
  wrap.appendChild(actions);

  // --- обработчики событий ---

  // переключение статуса через API (POST /api/toggle с JSON {id: ...})
  checkbox.addEventListener('click', async (e) => {
    e.preventDefault();
    // оптимистично меняем UI — пользователь видит мгновенный отклик
    task.Done = !task.Done;
    checkbox.textContent = task.Done ? '✔' : '';
    name.className = 'name' + (task.Done ? ' done' : '');

    try {
      await apiFetch('/api/toggle', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ id: task.ID })
      });
      showStatus('Статус обновлён');
    } catch (err) {
      // если ошибка, откатываем UI и показываем сообщение
      task.Done = !task.Done;
      checkbox.textContent = task.Done ? '✔' : '';
      name.className = 'name' + (task.Done ? ' done' : '');
      showStatus('Не удалось обновить статус: ' + err.message);
      console.error(err);
    }
  });

  // удаление задачи (DELETE /api/delete?id=...)
  del.addEventListener('click', async (e) => {
    e.preventDefault();
    if (!confirm('Удалить задачу?')) return;

    try {
      await apiFetch(`/api/delete?id=${task.ID}`, { method: 'DELETE' });
      wrap.remove(); // убрать элемент из DOM
      showStatus('Задача удалена');
    } catch (err) {
      showStatus('Ошибка удаления: ' + err.message);
      console.error(err);
    }
  });

  return wrap;
}

// Рендер всего списка задач: очищаем контейнер и добавляем карточки
function renderTasks(tasks) {
  tasksContainer.innerHTML = ''; // очистить
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

// Обработка отправки формы добавления задачи
addForm.addEventListener('submit', async (e) => {
  e.preventDefault(); // предотвратить перезагрузку страницы
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
    await loadTasks(); // перезагрузим список с сервера (надёжный способ синхронизироваться)
  } catch (err) {
    showStatus('Ошибка при добавлении: ' + err.message);
    console.error(err);
  }
});

// Загрузим задачи при старте страницы
loadTasks();
