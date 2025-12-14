const API_URL = 'http://localhost:8080';
let allMovies = [];
let searchTimeout;
let isAdminLoggedIn = false;


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
        showMessage('home-page', `Ошибка загрузки: ${error.message}`, 'error');
        console.error(error);
    }
}


async function loadMovieDetail(movieId) {
    try {
        // Загружаем основные данные фильма
        const response = await fetch(`${API_URL}/api/movies?id=${movieId}`);
        if (!response.ok) throw new Error('Movie not found');
        
        const movie = await response.json();
        
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
            showMessage('home-page', 
                `Фильм "${query}" не найден. Хотите парсить его?`, 
                'info'
            );
            renderMovies([]);
        } else {
            renderMovies(results);
        }
    } catch (error) {
        showMessage('home-page', `Ошибка поиска: ${error.message}`, 'error');
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
                        <p class="rating">⭐ ${movie.rating.toFixed(1)}</p>
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
    
    const descriptionId = `desc-${movie.id || 'unknown'}`;
    const MAX_LINES = 3;
    const isLongDescription = description.split('\n').length > MAX_LINES || description.length > 300;
    
    const descriptionHtml = isLongDescription 
        ? `
            <div class="detail-description">
                <div id="${descriptionId}" class="description-text collapsed" style="
                    display: -webkit-box;
                    -webkit-line-clamp: 3;
                    -webkit-box-orient: vertical;
                    overflow: hidden;
                    text-overflow: ellipsis;
                    transition: all 0.3s ease;
                ">
                    ${description}
                </div>
                <button class="btn-expand" id="${descriptionId}-btn" onclick="toggleDescription('${descriptionId}')" style="
                    margin-top: var(--spacing-sm);
                    padding: 4px 12px;
                    background: none;
                    border: 1px solid var(--color-accent);
                    color: var(--color-accent);
                    cursor: pointer;
                    border-radius: 4px;
                    font-size: 14px;
                    transition: all 0.2s ease;
                ">
                    Показать полностью ↓
                </button>
            </div>
        `
        : `
            <div class="detail-description">
                ${description}
            </div>
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
                    ${descriptionHtml}
                    <div class="detail-actions">
                        <button class="btn btn-primary" onclick="copyMovieInfo(${movie.id})">
                            📋 Скопировать информацию
                        </button>
                    </div>
                </div>
            </div>
            ${insightHtml}
            ${commentsHtml}
        </div>
    `;
}


function toggleDescription(descriptionId) {
    const element = document.getElementById(descriptionId);
    const button = document.getElementById(`${descriptionId}-btn`);
    
    if (element.classList.contains('collapsed')) {
        element.style.webkitLineClamp = 'unset';
        element.style.overflow = 'visible';
        element.classList.remove('collapsed');
        button.textContent = 'Скрыть ↑';
    } else {
        element.style.webkitLineClamp = '3';
        element.style.overflow = 'hidden';
        element.classList.add('collapsed');
        button.textContent = 'Показать полностью ↓';
    }
}


function copyMovieInfo(movieId) {
    const movie = allMovies.find(m => m.id === movieId);
    if (!movie) return;

    const text = `
${movie.title}

${movie.description || 'Описание отсутствует'}
    `.trim();

    navigator.clipboard.writeText(text).then(() => {
        showMessage('detail-page', '✓ Скопировано в буфер обмена', 'success');
    });
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
function showMessage(pageId, text, type = 'info') {
    const messageEl = document.getElementById(`${pageId}-message`) || 
                     document.querySelector(`#${pageId} .message-container`);
    
    if (!messageEl) return;

    messageEl.innerHTML = `<div class="message ${type}">${text}</div>`;
    setTimeout(() => clearMessages(`${pageId}-message`), 5000);
}


function clearMessages(elementId) {
    const el = document.getElementById(elementId);
    if (el) el.innerHTML = '';
}


// Инициализация
document.addEventListener('DOMContentLoaded', loadMovies);
