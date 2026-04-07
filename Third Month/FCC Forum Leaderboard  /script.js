const forumLatest = 'https://cdn.freecodecamp.org/curriculum/forum-latest/latest.json';
const forumTopicUrl = 'https://forum.freecodecamp.org/t/';
const postsContainer = document.getElementById('posts-container');

const categoryMap = {
  299: { name: 'Career Advice', className: 'career' },
  409: { name: 'Project Feedback', className: 'feedback' },
  417: { name: 'freeCodeCamp Support', className: 'support' },
  421: { name: 'JavaScript', className: 'javascript' },
  423: { name: 'HTML / CSS', className: 'html-css' },
  424: { name: 'Python', className: 'python' },
  432: { name: 'Backend Development', className: 'backend' },
  522: { name: 'General', className: 'general' },
  582: { name: 'Motivation', className: 'motivation' },
};

function getCategory(id) {
  return categoryMap[id] || { name: 'General', className: 'general' };
}

function timeAgo(dateStr) {
  const now = new Date();
  const past = new Date(dateStr);
  const diffMs = now - past;
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMins / 60);
  const diffDays = Math.floor(diffHours / 24);

  if (diffMins < 60) return `${diffMins}m ago`;
  if (diffHours < 24) return `${diffHours}h ago`;
  return `${diffDays}d ago`;
}

async function fetchAndRender() {
  try {
    const res = await fetch(forumLatest);
    if (!res.ok) throw new Error('Fetch failed');
    const data = await res.json();

    const { topic_list, users } = data;

    topic_list.topics.forEach((topic) => {
      const category = getCategory(topic.category_id);
      const avatarsHtml = topic.posters
        .slice(0, 5)
        .map((poster) => {
          const user = users.find((u) => u.id === poster.user_id);
          if (!user) return '';
          const src = user.avatar_template.replace('{size}', '30');
          return `<img src="${src}" alt="${user.username}" title="${user.username}" />`;
        })
        .join('');

      const row = document.createElement('tr');
      row.innerHTML = `
        <td>
          <a class="post-title" href="${forumTopicUrl}${topic.slug}/${topic.id}" target="_blank">
            ${topic.title}
          </a>
          <a class="category ${category.className}" href="https://forum.freecodecamp.org/c/${topic.category_id}" target="_blank">
            ${category.name}
          </a>
        </td>
        <td>
          <div class="avatar-container">${avatarsHtml}</div>
        </td>
        <td>${topic.posts_count - 1}</td>
        <td>${topic.views}</td>
        <td>${timeAgo(topic.last_posted_at)}</td>
      `;
      postsContainer.appendChild(row);
    });
  } catch (err) {
    postsContainer.innerHTML =
      '<tr><td colspan="5" style="text-align:center;padding:20px;">Failed to load topics. Please try again later.</td></tr>';
  }
}

fetchAndRender();
