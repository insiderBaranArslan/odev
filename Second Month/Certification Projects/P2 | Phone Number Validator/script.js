const userInput = document.getElementById("user-input");
const checkBtn = document.getElementById("check-btn");
const clearBtn = document.getElementById("clear-btn");
const resultsDiv = document.getElementById("results-div");

// Valid US phone number regex
const phoneRegex = /^(1\s?)?(\(\d{3}\)|\d{3})[\s\-]?\d{3}[\s\-]?\d{4}$/;

const checkPhone = () => {
  const inputVal = userInput.value;

  if (!inputVal) {
    alert("Please provide a phone number");
    return;
  }

  const isValid = phoneRegex.test(inputVal);
  const resultItem = document.createElement("p");
  resultItem.classList.add("result-item", isValid ? "valid" : "invalid");
  resultItem.textContent = isValid
    ? `Valid US number: ${inputVal}`
    : `Invalid US number: ${inputVal}`;

  resultsDiv.appendChild(resultItem);
  userInput.value = "";
};

checkBtn.addEventListener("click", checkPhone);

userInput.addEventListener("keydown", (e) => {
  if (e.key === "Enter") checkPhone();
});

clearBtn.addEventListener("click", () => {
  resultsDiv.innerHTML = "";
});
