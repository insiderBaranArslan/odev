const textInput = document.getElementById("text-input");
const checkBtn = document.getElementById("check-btn");
const result = document.getElementById("result");

const checkPalindrome = () => {
  const inputVal = textInput.value;

  if (!inputVal) {
    alert("Please input a value");
    return;
  }

  // Remove non-alphanumeric characters and lowercase
  const cleaned = inputVal.replace(/[^a-zA-Z0-9]/g, "").toLowerCase();
  const reversed = cleaned.split("").reverse().join("");
  const isPalindrome = cleaned === reversed;

  result.textContent = isPalindrome
    ? `${inputVal} is a palindrome`
    : `${inputVal} is not a palindrome`;

  result.className = isPalindrome ? "is-palindrome" : "not-palindrome";
};

checkBtn.addEventListener("click", checkPalindrome);

textInput.addEventListener("keydown", (e) => {
  if (e.key === "Enter") checkPalindrome();
});
