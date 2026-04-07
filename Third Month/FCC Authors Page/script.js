const authorContainer = document.getElementById('author-container');
const loadMoreBtn = document.getElementById('load-more-btn');

const AUTHORS_PER_PAGE = 8;
let currentPage = 0;
let allAuthors = [];

async function fetchAuthors() {
  try {
    const response = await fetch(
      'https://cdn.freecodecamp.org/curriculum/news-author-page/authors.json'
    );
    if (!response.ok) throw new Error('Failed to fetch authors');
    allAuthors = await response.json();
    displayAuthors();
  } catch (err) {
    authorContainer.innerHTML =
      '<p class="error-msg">There was an error loading the authors. Please try again later.</p>';
    loadMoreBtn.style.display = 'none';
  }
}

function displayAuthors() {
  const start = currentPage * AUTHORS_PER_PAGE;
  const end = start + AUTHORS_PER_PAGE;
  const pageAuthors = allAuthors.slice(start, end);

  pageAuthors.forEach(({ author, image, url, bio }) => {
    const card = document.createElement('div');
    card.classList.add('user-card');
    card.innerHTML = `
      <div class="purple-divider"></div>
      <img class="user-img" src="${image}" alt="${author}" />
      <p class="author-name"><a href="${url}" target="_blank">${author}</a></p>
      <p class="bio">${bio.slice(0, 100)}${bio.length > 100 ? '...' : ''}</p>
    `;
    authorContainer.appendChild(card);
  });

  currentPage++;

  if (currentPage * AUTHORS_PER_PAGE >= allAuthors.length) {
    loadMoreBtn.style.display = 'none';
  }
}

loadMoreBtn.addEventListener('click', displayAuthors);

fetchAuthors();
