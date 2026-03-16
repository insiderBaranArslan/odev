// Get the current date paragraph and dropdown elements
const currentDateParagraph = document.getElementById("current-date");
const dateOptionsSelectElement = document.getElementById("date-options");

// Create a Date object for today
const date = new Date();

// Extract date parts
const day = date.getDate();
const month = date.getMonth() + 1;
const year = date.getFullYear();
const hours = date.getHours();
const minutes = date.getMinutes();

// Format the date as dd-mm-yyyy by default
currentDateParagraph.textContent = `${day}-${month}-${year}`;

// Listen for dropdown changes and update the displayed date format
dateOptionsSelectElement.addEventListener("change", () => {
  switch (dateOptionsSelectElement.value) {
    case "yyyy-mm-dd":
      currentDateParagraph.textContent = `${year}-${month}-${day}`;
      break;
    case "mm-dd-yyyy-h-mm":
      currentDateParagraph.textContent = `${month}-${day}-${year}-${hours}h ${minutes}m`;
      break;
    default:
      currentDateParagraph.textContent = `${day}-${month}-${year}`;
  }
});
