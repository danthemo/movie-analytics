const API_URL = window.location.origin;
let allMovies = [];
let searchTimeout;
let isAdminLoggedIn = false;
let currentMovie = null;


// Навигация
function showPage(pageId) {
    document.querySelectorAll('.page').forEach(p => p.classList.remove('active'));
    document.getElementById(pageId).classList.add('active');


    // Активный пункт навигации
    document.querySelectorAll('nav button').forEach(b => b.classList.remove('active'));
    if (pageId === 'home-page') document.getElementById('nav-home').classList.add('active');
    if (pageId === 'scrape-page') document.getElementById('nav-scrape').classList.add('active');
    if (pageId === 'admin-page') document.getElementById('nav-admin').classList.add('active');
}


function showHome() {
    showPage('home-page');
    loadMovies();
}


function showScrape() {
    showPage('scrape-page');
    clearMessages('scrape-message');
}


function showDetail(movieId) {
    showPage('detail-page');
    loadMovieDetail(movieId);
}


function showAdmin() {
    if (!isAdminLoggedIn) {
        showPage('admin-login-page');
    } else {
        showPage('admin-page');
        loadAdminStats();
        loadAdminMovies();
    }
}


// Вызовы АПИ
async function loadMovies() {
    try {
        const response = await fetch(`${API_URL}/api/movies`);
        if (!response.ok) throw new Error('Failed to load movies');
        
        allMovies = await response.json();
        renderMovies(allMovies);
    } catch (error) {
        showMessage('search-message', `Ошибка загрузки: ${error.message}`, 'error');
        console.error(error);
    }
}


async function loadMovieDetail(movieId) {
    try {
        // Загружаем основные данные фильма
        const response = await fetch(`${API_URL}/api/movies?id=${movieId}`);
        if (!response.ok) throw new Error('Movie not found');
        
        const movie = await response.json();
        currentMovie = movie;
        
        // Загружаем insights отдельно
        try {
            const insightResponse = await fetch(`${API_URL}/api/movies/insights?id=${movieId}`);
            if (insightResponse.ok) {
                const insight = await insightResponse.json();
                movie.insight = insight;
            }
        } catch (e) {
            console.log('Insights not yet available');
        }
        
        renderMovieDetail(movie);
    } catch (error) {
        document.getElementById('detail-content').innerHTML = 
            `<div class="empty-state"><h2>Ошибка загрузки</h2><p>${error.message}</p></div>`;
        console.error(error);
    }
}


async function searchMovies() {
    const query = document.getElementById('search-input').value.trim();
    if (!query) {
        loadMovies();
        return;
    }


    try {
        clearMessages('search-message');
        const response = await fetch(`${API_URL}/api/search?query=${encodeURIComponent(query)}`);
        if (!response.ok) throw new Error('Search failed');
        
        const results = await response.json();
        
        if (results.length === 0) {
            showMessage('search-message', 
                `Фильм "${query}" не найден. Хотите парсить его?`, 
                'info'
            );
            renderMovies([]);
        } else {
            renderMovies(results);
        }
    } catch (error) {
        showMessage('search-message', `Ошибка поиска: ${error.message}`, 'error');
    }
}


function debounceSearch() {
    clearTimeout(searchTimeout);
    searchTimeout = setTimeout(searchMovies, 500);
}


async function submitScrape(event) {
    event.preventDefault();
    
    const title = document.getElementById('movie-title').value.trim();


    if (!title) {
        showToast('Введите название фильма', 'error');
        return;
    }


    try {
        clearMessages('scrape-message');
        const msgEl = document.getElementById('scrape-message');
        msgEl.innerHTML = '<div class="loading"><div class="spinner"></div>Парсим фильм...</div>';


        const response = await fetch(
            `${API_URL}/api/scrape?query=${encodeURIComponent(title)}`,
            {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({})
            }
        );


        if (!response.ok) {
            let errorText = 'Ошибка при парсинге фильма';
            
            try {
                const data = await response.json();
                if (data.error) errorText = data.error;
            } catch (e) {
                try {
                    const text = await response.text();
                    if (text) errorText = text;
                } catch (e2) {
                    errorText = `Ошибка сервера (${response.status})`;
                }
            }
            
            showToast(errorText, 'error');
            showHome();
            return;
        }


        const result = await response.json();
        showToast(`✓ Фильм "${result.title}" успешно добавлен!`, 'success');
        event.target.reset();
        
        setTimeout(() => {
            showDetail(result.id);
        }, 1500);
    } catch (error) {
        showToast(error.message, 'error');
    }
}


// функици администратора
function submitAdminLogin(event) {
    event.preventDefault();
    
    const password = document.getElementById('admin-password').value;
    
    if (password === '1234') {
        isAdminLoggedIn = true;
        document.getElementById('admin-password').value = '';
        showPage('admin-page');
        loadAdminStats();
        loadAdminMovies();
        showToast('✓ Вы вошли в админ панель', 'success');
    } else {
        showToast('❌ Неверный пароль', 'error');
    }
}


function logoutAdmin() {
    isAdminLoggedIn = false;
    showHome();
    showToast('Вы вышли из админ панели', 'info');
}


function loadAdminStats() {
    try {
        const movies = allMovies || [];
        document.getElementById('stat-movies').textContent = movies.length;
        
        let totalComments = 0;
        movies.forEach(m => {
            if (m.comments) totalComments += m.comments.length;
        });
        document.getElementById('stat-comments').textContent = totalComments;
    } catch (error) {
        console.error(error);
    }
}


async function loadAdminMovies() {
    try {
        const response = await fetch(`${API_URL}/api/movies`);
        if (!response.ok) throw new Error('Failed to load movies');
        
        const movies = await response.json();
        
        const html = movies.map(movie => `
            <div style="background-color: var(--color-bg); padding: var(--spacing-md); border-radius: var(--radius); margin-bottom: var(--spacing-md); display: flex; justify-content: space-between; align-items: center;">
                <div>
                    <div style="font-weight: 600;">${movie.title}</div>
                    <div style="color: var(--color-text-secondary); font-size: 12px;">
                        ${movie.comments?.length || 0} комментариев
                    </div>
                </div>
                <button class="btn btn-secondary" onclick="deleteMovieAdmin(${movie.id})">🗑️ Удалить</button>
            </div>
        `).join('');
        
        const container = document.getElementById('admin-movies-list');
        if (html) {
            container.innerHTML = html;
        } else {
            container.innerHTML = '<div style="color: var(--color-text-secondary);">Нет фильмов</div>';
        }
    } catch (error) {
        console.error(error);
    }
}


async function deleteMovieAdmin(movieId) {
    if (!confirm('Точно удалить фильм?')) return;
    
    try {
        const response = await fetch(`${API_URL}/api/movies?id=${movieId}`, {
            method: 'DELETE'
        });
        
        if (response.ok) {
            showToast('✓ Фильм удален', 'success');
            loadAdminMovies();
            loadAdminStats();
            loadMovies();
        } else {
            showToast('Ошибка удаления', 'error');
        }
    } catch (error) {
        showToast('Ошибка удаления', 'error');
        console.error(error);
    }
}


// Рендер
function renderMovies(movies) {
    const container = document.getElementById('movies-container');
    
    if (!movies || movies.length === 0) {
        container.innerHTML = `
            <div class="empty-state">
                <h2>Фильмы не найдены</h2>
                <p>Попробуйте другой поиск или добавьте новый фильм</p>
            </div>
        `;
        return;
    }


    container.innerHTML = `
        <div class="movies-grid">
            ${movies.map(movie => `
                <div class="movie-card" onclick="showDetail(${movie.id})">
                    <div class="movie-poster">
                        ${movie.poster_url 
                            ? `<img src="${movie.poster_url}" alt="${movie.title}">`
                            : '🎬 Нет постера'
                        }
                    </div>
                    <div class="movie-info">
                        <div class="movie-title">${movie.title}</div>
                        <p class="rating">${formatMovieRating(movie)}</p>
                    </div>
                </div>
            `).join('')}
        </div>
    `;
}


function renderMovieDetail(movie) {
    const title = movie?.title || 'Неизвестный фильм';
    const description = movie?.description || 'Описание отсутствует';
    const posterUrl = movie?.poster_url;
    const comments = Array.isArray(movie?.comments) ? movie.comments : [];
    const metadata = buildMovieMetadata(movie);
    
    const descriptionId = `desc-${movie.id || 'unknown'}`;
    const normalizedDescription = normalizeDescriptionText(description);
    const formattedDescription = formatDescriptionHtml(normalizedDescription);
    const isLongDescription = normalizedDescription.length > 700 || normalizedDescription.split('\n').length > 10;
    
    const descriptionHtml = isLongDescription 
        ? `
            <section class="detail-description-section">
                <div class="detail-description-header">
                    <h3>Описание</h3>
                </div>
                <div id="${descriptionId}" class="detail-description-content collapsed" data-collapsed="true">
                    ${formattedDescription}
                </div>
                <button class="btn-expand" id="${descriptionId}-btn" onclick="toggleDescription('${descriptionId}')">
                    Показать полностью ↓
                </button>
            </section>
        `
        : `
            <section class="detail-description-section">
                <div class="detail-description-header">
                    <h3>Описание</h3>
                </div>
                <div class="detail-description-content">
                    ${formattedDescription}
                </div>
            </section>
        `;
    
    // Комментарии
    let commentsHtml = '';
    if (comments.length > 0) {
        commentsHtml = `
            <div style="margin-top: var(--spacing-lg); padding-top: var(--spacing-lg); border-top: 1px solid var(--color-border);">
                <h3 style="margin-bottom: var(--spacing-md); font-size: 20px;">💬 Комментарии (${comments.length})</h3>
                <div style="display: flex; flex-direction: column; gap: var(--spacing-md);">
                    ${comments.map(comment => {
                        const date = new Date(comment.scraped_at);
                        const dateStr = date.toLocaleDateString('ru-RU', { 
                            year: 'numeric', 
                            month: 'long', 
                            day: 'numeric' 
                        });
                        return `
                            <div style="background-color: var(--color-bg); padding: var(--spacing-md); border-radius: var(--radius); border-left: 3px solid var(--color-accent);">
                                <div style="display: flex; justify-content: space-between; margin-bottom: var(--spacing-sm);">
                                    <strong style="color: var(--color-accent);">${comment.author || 'Аноним'}</strong>
                                    <span style="color: var(--color-text-secondary); font-size: 12px;">${dateStr}</span>
                                </div>
                                <p style="margin: 0; line-height: 1.6; color: var(--color-text);">${comment.text || ''}</p>
                                <span style="display: inline-block; margin-top: var(--spacing-sm); font-size: 11px; color: var(--color-text-secondary); background-color: var(--color-surface); padding: 2px 8px; border-radius: var(--radius);">
                                    ${comment.source || 'Неизвестный источник'}
                                </span>
                            </div>
                        `;
                    }).join('')}
                </div>
            </div>
        `;
    } else {
        commentsHtml = `
            <div style="margin-top: var(--spacing-lg); padding-top: var(--spacing-lg); border-top: 1px solid var(--color-border);">
                <h3 style="margin-bottom: var(--spacing-md); font-size: 20px;">💬 Комментарии</h3>
                <div class="empty-state">
                    <h2>Комментариев пока нет</h2>
                    <p>Для этого фильма еще не удалось получить отзывы, поэтому AI-анализ и рейтинг могут отсутствовать.</p>
                </div>
            </div>
        `;
    }
    
    // ИИ сводка
    let insightHtml = '';
    if (movie?.insight && movie.insight.summary) {
        const rating = movie.insight.rating || 0;

        insightHtml = `
            <div style="margin-top: var(--spacing-lg); padding-top: var(--spacing-lg); border-top: 1px solid var(--color-border);">
                <h3 style="margin-bottom: var(--spacing-md); font-size: 20px;">🤖 AI Анализ</h3>
                
                <div style="background: linear-gradient(135deg, rgba(33, 128, 168, 0.1), rgba(50, 184, 198, 0.1)); padding: var(--spacing-md); border-radius: var(--radius); margin-bottom: var(--spacing-md); border-left: 4px solid var(--color-accent);">
                    <div style="display: flex; justify-content: space-between; align-items: center;">
                        <span style="font-weight: 600;">Оценка:</span>
                        <span style="color: var(--color-accent); font-weight: 600; font-size: 24px;">⭐ ${rating.toFixed(1)}/10</span>
                    </div>
                </div>
                
                <div style="background-color: var(--color-surface); padding: var(--spacing-md); border-radius: var(--radius);">
                    <h4 style="margin-bottom: var(--spacing-sm); color: var(--color-accent);">📝 Сводка от ИИ</h4>
                    <p style="margin: 0; line-height: 1.6; color: var(--color-text);">${movie.insight.summary}</p>
                </div>
            </div>
        `;
    } else {
        insightHtml = `
            <div style="margin-top: var(--spacing-lg); padding-top: var(--spacing-lg); border-top: 1px solid var(--color-border);">
                <h3 style="margin-bottom: var(--spacing-md); font-size: 20px;">🤖 AI Анализ</h3>
                <div class="empty-state">
                    <h2>Анализ пока недоступен</h2>
                    <p>Недостаточно данных для оценки фильма или AI-сервис временно не смог обработать отзывы.</p>
                </div>
            </div>
        `;
    }
    
    // Рендерим всё вместе
    document.getElementById('detail-content').innerHTML = `
        <div class="movie-detail">
            <div class="detail-header">
                <div class="detail-poster">
                    ${posterUrl 
                        ? `<img src="${posterUrl}" alt="${title}">`
                        : '🎬'
                    }
                </div>
                <div class="detail-info">
                    <h1>${title}</h1>
                    ${metadata ? `<div class="detail-meta">${metadata}</div>` : ''}
                </div>
            </div>
            ${descriptionHtml}
            ${insightHtml}
            ${commentsHtml}
        </div>
    `;
}


function toggleDescription(descriptionId) {
    const element = document.getElementById(descriptionId);
    const button = document.getElementById(`${descriptionId}-btn`);
    if (!element || !button) return;
    
    const isCollapsed = element.dataset.collapsed !== 'false';

    if (isCollapsed) {
        element.classList.remove('collapsed');
        element.dataset.collapsed = 'false';
        button.textContent = 'Скрыть ↑';
    } else {
        element.classList.add('collapsed');
        element.dataset.collapsed = 'true';
        button.textContent = 'Показать полностью ↓';
    }
}


// Уведомления
function showToast(text, type = 'info') {
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.innerHTML = `
        <div class="toast-content">${text}</div>
        <button class="toast-close" onclick="this.parentElement.remove()">✕</button>
    `;
    document.body.appendChild(toast);
    
    setTimeout(() => {
        toast.style.animation = 'slideOut 300ms ease-in forwards';
        setTimeout(() => toast.remove(), 300);
    }, 4000);
}


// Сообщения на страницах
function showMessage(elementId, text, type = 'info') {
    const messageEl = document.getElementById(elementId);
    
    if (!messageEl) return;

    messageEl.innerHTML = `<div class="message ${type}">${text}</div>`;
    setTimeout(() => clearMessages(elementId), 5000);
}


function clearMessages(elementId) {
    const el = document.getElementById(elementId);
    if (el) el.innerHTML = '';
}


// Инициализация
document.addEventListener('DOMContentLoaded', loadMovies);


function formatMovieRating(movie) {
    const rating = Number(movie?.rating);
    const commentsCount = Array.isArray(movie?.comments) ? movie.comments.length : 0;

    if (Number.isFinite(rating) && rating > 0) {
        return `⭐ ${rating.toFixed(1)}`;
    }

    if (commentsCount === 0) {
        return 'Не оценено';
    }

    return 'Оценка недоступна';
}


function normalizeDescriptionText(rawText) {
    if (!rawText) return 'Описание отсутствует';

    let text = rawText.replace(/\r/g, '').trim();
    const sectionTitles = ['Сюжет', 'Причины посмотреть', 'Интересные факты', 'Осторожно, спойлеры!'];

    for (const title of sectionTitles) {
        const escapedTitle = title.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
        text = text.replace(new RegExp(`([^\\n])(${escapedTitle})`, 'g'), `$1\n\n$2`);
    }

    text = text.replace(/▪\s*/g, '\n▪ ');
    text = text.replace(/\n{3,}/g, '\n\n');
    return text.trim();
}


function formatDescriptionHtml(text) {
    const safeText = escapeHtml(text);
    const sections = safeText.split(/\n{2,}/).map(section => section.trim()).filter(Boolean);

    return sections.map(section => renderDescriptionSection(section)).join('');
}


function renderDescriptionSection(section) {
    const titles = ['Сюжет', 'Причины посмотреть', 'Интересные факты', 'Осторожно, спойлеры!'];
    const matchedTitle = titles.find(title => section.startsWith(title));

    if (!matchedTitle) {
        return `<p>${section.replace(/\n/g, '<br>')}</p>`;
    }

    const body = section.slice(matchedTitle.length).trim();
    const lines = body.split('\n').map(line => line.trim()).filter(Boolean);
    const bulletLines = lines.filter(line => line.startsWith('▪'));
    const textLines = lines.filter(line => !line.startsWith('▪'));

    const paragraphs = textLines.length > 0
        ? `<p>${textLines.join('<br>')}</p>`
        : '';

    const bullets = bulletLines.length > 0
        ? `
            <ul class="description-bullets">
                ${bulletLines.map(line => `<li>${line.replace(/^▪\s*/, '')}</li>`).join('')}
            </ul>
        `
        : '';

    return `
        <div class="description-section">
            <h4>${matchedTitle}</h4>
            ${paragraphs}
            ${bullets}
        </div>
    `;
}


function escapeHtml(value) {
    return String(value)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;');
}


function buildMovieMetadata(movie) {
    const parts = [];

    if (Number(movie?.year) > 0) {
        parts.push(`Год: ${movie.year}`);
    }

    if (movie?.directors) {
        parts.push(`Режиссёр: ${escapeHtml(movie.directors)}`);
    }

    if (movie?.actors) {
        parts.push(`Актёры: ${escapeHtml(movie.actors)}`);
    }

    return parts.join(' • ');
}
