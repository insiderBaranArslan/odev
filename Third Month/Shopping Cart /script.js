const cartContainer = document.getElementById('cart-container');
const productsContainer = document.getElementById('products-container');
const dessertCards = document.getElementById('dessert-card-container');

const cartBtn = document.getElementById('cart-btn');
const clearCartBtn = document.getElementById('clear-cart-btn');
const showHideCart = document.getElementById('show-hide-cart');

const totalItemsEl = document.getElementById('total-items');
const subtotalEl = document.getElementById('subtotal');
const taxesEl = document.getElementById('taxes');
const totalEl = document.getElementById('total');

const TAX_RATE = 0.1;

const desserts = [
  { id: 1, name: 'Chocolate Cake', price: 12.99, category: 'Cakes' },
  { id: 2, name: 'Strawberry Cheesecake', price: 9.99, category: 'Cakes' },
  { id: 3, name: 'Vanilla Ice Cream', price: 4.99, category: 'Ice Cream' },
  { id: 4, name: 'Chocolate Brownie', price: 3.99, category: 'Cookies' },
  { id: 5, name: 'Creme Brulee', price: 7.99, category: 'Custards' },
  { id: 6, name: 'Tiramisu', price: 10.99, category: 'Cakes' },
];

// Cart state: { [dessert.id]: { dessert, count } }
const cart = {};

// Render dessert cards
desserts.forEach((dessert) => {
  const card = document.createElement('div');
  card.classList.add('dessert-card');
  card.innerHTML = `
    <h2>${dessert.name}</h2>
    <p class="product-category">${dessert.category}</p>
    <p class="dessert-price">$${dessert.price.toFixed(2)}</p>
    <button class="btn add-to-cart-btn" data-id="${dessert.id}">Add to Cart</button>
  `;
  dessertCards.appendChild(card);
});

// Add to cart
dessertCards.addEventListener('click', (e) => {
  if (!e.target.classList.contains('add-to-cart-btn')) return;

  const id = Number(e.target.dataset.id);
  const dessert = desserts.find((d) => d.id === id);
  if (!dessert) return;

  if (cart[id]) {
    cart[id].count += 1;
  } else {
    cart[id] = { dessert, count: 1 };
  }

  updateCart();
});

// Clear cart
clearCartBtn.addEventListener('click', () => {
  Object.keys(cart).forEach((key) => delete cart[key]);
  updateCart();
});

// Show/hide cart
cartBtn.addEventListener('click', () => {
  const isHidden = cartContainer.style.display === 'none' || cartContainer.style.display === '';
  cartContainer.style.display = isHidden ? 'block' : 'none';
  showHideCart.textContent = isHidden ? 'Hide' : 'Show';
});

function updateCart() {
  productsContainer.innerHTML = '';

  let totalItems = 0;
  let subtotal = 0;

  Object.values(cart).forEach(({ dessert, count }) => {
    totalItems += count;
    subtotal += dessert.price * count;

    const productEl = document.createElement('div');
    productEl.classList.add('product');
    productEl.innerHTML = `
      <p><span class="product-count">${count}x</span>${dessert.name}</p>
      <p>$${(dessert.price * count).toFixed(2)}</p>
    `;
    productsContainer.appendChild(productEl);
  });

  const taxes = subtotal * TAX_RATE;
  const total = subtotal + taxes;

  totalItemsEl.textContent = totalItems;
  subtotalEl.textContent = `$${subtotal.toFixed(2)}`;
  taxesEl.textContent = `$${taxes.toFixed(2)}`;
  totalEl.textContent = `$${total.toFixed(2)}`;
}
