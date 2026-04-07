const searchInput = document.getElementById('search-input');
const searchButton = document.getElementById('search-button');
const card = document.getElementById('card');

const creatureNameEl = document.getElementById('creature-name');
const creatureIdEl = document.getElementById('creature-id');
const weightEl = document.getElementById('weight');
const heightEl = document.getElementById('height');
const typesEl = document.getElementById('types');
const hpEl = document.getElementById('hp');
const attackEl = document.getElementById('attack');
const defenseEl = document.getElementById('defense');
const specialAttackEl = document.getElementById('special-attack');
const specialDefenseEl = document.getElementById('special-defense');
const speedEl = document.getElementById('speed');
const spriteEl = document.getElementById('creature-sprite');

const API_URL = 'https://rpg-creature-api.freecodecamp.rocks/api/creature';

searchButton.addEventListener('click', searchCreature);
searchInput.addEventListener('keydown', (e) => {
  if (e.key === 'Enter') searchCreature();
});

async function searchCreature() {
  const query = searchInput.value.trim();
  if (!query) return;

  try {
    const response = await fetch(`${API_URL}/${query.toLowerCase()}`);

    if (!response.ok) {
      alert('Creature not found');
      return;
    }

    const data = await response.json();

    if (!data || data.error) {
      alert('Creature not found');
      return;
    }

    displayCreature(data);
  } catch (err) {
    alert('Creature not found');
  }
}

function displayCreature(data) {
  creatureNameEl.textContent = data.name.toUpperCase();
  creatureIdEl.textContent = `#${data.id}`;
  weightEl.textContent = `Weight: ${data.weight}`;
  heightEl.textContent = `Height: ${data.height}`;

  typesEl.innerHTML = '';
  data.types.forEach((typeObj) => {
    const span = document.createElement('span');
    span.textContent = typeObj.name.toUpperCase();
    typesEl.appendChild(span);
  });

  const stats = data.stats;
  hpEl.textContent = getStat(stats, 'hp');
  attackEl.textContent = getStat(stats, 'attack');
  defenseEl.textContent = getStat(stats, 'defense');
  specialAttackEl.textContent = getStat(stats, 'special-attack');
  specialDefenseEl.textContent = getStat(stats, 'special-defense');
  speedEl.textContent = getStat(stats, 'speed');

  if (data.sprites && data.sprites.front_default) {
    spriteEl.src = data.sprites.front_default;
    spriteEl.alt = data.name;
    spriteEl.style.display = 'block';
  } else {
    spriteEl.style.display = 'none';
  }

  card.classList.add('visible');
}

function getStat(stats, statName) {
  const found = stats.find((s) => s.name === statName);
  return found ? found.base_stat : '-';
}
