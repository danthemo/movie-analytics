from playwright.sync_api import sync_playwright
import re
from urllib.parse import quote

def clean_text(text):
    return re.sub(r'\s+', ' ', text).strip()

def parse_okko_film_info(query):
    query_encoded = quote(query)
    search_url = f"https://okko.tv/search/{query_encoded}"

    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        page.goto(search_url)
        
        page.wait_for_selector('a.vj8iwpkR', timeout=10000)
        
        # Находим ссылку на фильм
        links = page.query_selector_all('div.Oce5ukhe a.vj8iwpkR')
        film_link = None
        for l in links:
            text = clean_text(l.inner_text().lower())
            if query.lower() in text:
                film_link = l.get_attribute("href")
                break
        
        if not film_link:
            browser.close()
            return {}
        
        film_url = f"https://okko.tv{film_link}"
        page.goto(film_url)
        page.wait_for_selector('h1', timeout=10000)
        
        film_info = {}

        # Название
        title_element = page.query_selector('h1')
        if title_element:
            title = clean_text(title_element.inner_text())
            title = title.replace("Смотреть онлайн", "").strip()
            film_info['title'] = title

        # Описание
        description_element = page.query_selector('span[test-id="content_description"]')
        film_info['description'] = clean_text(description_element.inner_text()) if description_element else ""

        # Год
        year_element = page.query_selector('span[test-id="meta_release_date"]')
        film_info['year'] = clean_text(year_element.inner_text()) if year_element else ""

        # Режиссёры
        film_info['directors'] = []
        director_block = page.query_selector('div[test-id="meta_director"]')
        if director_block:
            director_links = director_block.query_selector_all('a.aPyppOuR')
            film_info['directors'] = [clean_text(a.inner_text()) for a in director_links]

        # Актёры
        film_info['actors'] = []
        actors_block = page.query_selector('div[test-id="meta_actors"]')
        if actors_block:
            actor_links = actors_block.query_selector_all('a.aPyppOuR')
            film_info['actors'] = [clean_text(a.inner_text()) for a in actor_links]

        browser.close()
        return film_info

# Пример использования
# info = parse_okko_film_info("мстители")
# print(info)


def parse_okko_reviews(query):
    query_encoded = quote(query)
    search_url = f"https://okko.tv/search/{query_encoded}"

    with sync_playwright() as p:
        browser = p.chromium.launch(headless=True)
        page = browser.new_page()
        page.goto(search_url)
        
        # Ждем результатов поиска
        page.wait_for_selector('a.vj8iwpkR', timeout=10000)
        
        # Ссылка на фильм
        links = page.query_selector_all('div.Oce5ukhe a.vj8iwpkR')
        film_link = None
        for l in links:
            text = clean_text(l.inner_text().lower())
            if query.lower() in text:
                film_link = l.get_attribute("href")
                break
        
        if not film_link:
            browser.close()
            return []  # Фильм не найден
        
        # Переходим на страницу фильма
        film_url = f"https://okko.tv{film_link}/reviews"
        page.goto(film_url)
        page.wait_for_selector('div.GLR27TKM', timeout=10000)
        
        # Сбор отзывов
        review_elements = page.query_selector_all('div.GLR27TKM.bUhuFIH2.KaAZjIfH')
        reviews = [clean_text(r.inner_text()) for r in review_elements if clean_text(r.inner_text())]

        browser.close()
        return reviews

# # Пример использования
# reviews = parse_okko_reviews("мстители")
# print(reviews)
