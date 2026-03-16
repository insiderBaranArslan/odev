// Get form and dialog elements
const taskForm = document.getElementById("task-form");
const confirmCloseDialog = document.getElementById("confirm-close-dialog");
const openTaskFormBtn = document.getElementById("open-task-form-btn");
const closeTaskFormBtn = document.getElementById("close-task-form-btn");
const addOrUpdateTaskBtn = document.getElementById("add-or-update-task-btn");
const cancelBtn = document.getElementById("cancel-btn");
const discardBtn = document.getElementById("discard-btn");
const tasksContainer = document.getElementById("tasks-container");
const titleInput = document.getElementById("title-input");
const dateInput = document.getElementById("date-input");
const descriptionInput = document.getElementById("description-input");

// Load tasks from localStorage, or start with empty array
const taskData = JSON.parse(localStorage.getItem("data")) || [];

// Track which task is being edited (null = adding new)
let currentTask = {};

// Render all tasks to the DOM
const renderTasksToDOM = () => {
  tasksContainer.innerHTML = "";

  taskData.forEach(({ id, title, date, description }) => {
    tasksContainer.innerHTML += `
      <div class="task" id="${id}">
        <p><strong>Title:</strong> ${title}</p>
        <p><strong>Date:</strong> ${date}</p>
        <p><strong>Description:</strong> ${description}</p>
        <button onclick="editTask(this)" type="button" class="btn">Edit</button>
        <button onclick="deleteTask(this)" type="button" class="btn">Delete</button>
      </div>
    `;
  });
};

// Check if any input has a value (to know if we should show discard dialog)
const hasFormInput = () =>
  titleInput.value || dateInput.value || descriptionInput.value;

// Open task form
openTaskFormBtn.addEventListener("click", () => {
  taskForm.classList.toggle("hidden");
});

// Close task form: if there is unsaved input, show confirm dialog
closeTaskFormBtn.addEventListener("click", () => {
  if (hasFormInput()) {
    confirmCloseDialog.showModal();
  } else {
    resetForm();
  }
});

// Cancel discard — close the dialog, keep the form open
cancelBtn.addEventListener("click", () => {
  confirmCloseDialog.close();
});

// Confirm discard — close dialog and reset form
discardBtn.addEventListener("click", () => {
  confirmCloseDialog.close();
  resetForm();
});

// Add or update task on form submit
taskForm.addEventListener("submit", (e) => {
  e.preventDefault();

  const dataArrIndex = taskData.findIndex((item) => item.id === currentTask.id);

  const taskObj = {
    id: currentTask.id || `${titleInput.value.toLowerCase().split(" ").join("-")}-${Date.now()}`,
    title: titleInput.value,
    date: dateInput.value,
    description: descriptionInput.value,
  };

  if (dataArrIndex === -1) {
    // New task
    taskData.unshift(taskObj);
  } else {
    // Update existing task
    taskData[dataArrIndex] = taskObj;
  }

  localStorage.setItem("data", JSON.stringify(taskData));
  renderTasksToDOM();
  resetForm();
});

// Edit a task — populate form with task data
const editTask = (buttonEl) => {
  const dataArrIndex = taskData.findIndex(
    (item) => item.id === buttonEl.parentElement.id
  );

  currentTask = taskData[dataArrIndex];

  titleInput.value = currentTask.title;
  dateInput.value = currentTask.date;
  descriptionInput.value = currentTask.description;

  addOrUpdateTaskBtn.textContent = "Update Task";
  taskForm.classList.remove("hidden");
};

// Delete a task
const deleteTask = (buttonEl) => {
  const dataArrIndex = taskData.findIndex(
    (item) => item.id === buttonEl.parentElement.id
  );

  buttonEl.parentElement.remove();
  taskData.splice(dataArrIndex, 1);
  localStorage.setItem("data", JSON.stringify(taskData));
};

// Reset form to initial state
const resetForm = () => {
  titleInput.value = "";
  dateInput.value = "";
  descriptionInput.value = "";
  taskForm.classList.add("hidden");
  addOrUpdateTaskBtn.textContent = "Add Task";
  currentTask = {};
};

// Render tasks on page load
renderTasksToDOM();
