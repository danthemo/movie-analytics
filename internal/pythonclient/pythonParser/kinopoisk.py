import requests
from bs4 import BeautifulSoup as bs
import re
import time

def parse_kp_reviews(raw_query):
    # Начальные значения
    session = requests.Session()
    session.headers.update({
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
        'Accept': 'text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8',
        'Accept-Language': 'ru-RU,ru;q=0.9,en;q=0.8',
        'Accept-Encoding': 'gzip, deflate, br',
        'Connection': 'keep-alive',
        'Upgrade-Insecure-Requests': '1',
        'Sec-Fetch-Dest': 'document',
        'Sec-Fetch-Mode': 'navigate',
        'Sec-Fetch-Site': 'same-origin',
        'Cache-Control': 'max-age=0',
        'Referer': 'https://www.kinopoisk.ru/',
    })

    query = raw_query.strip().replace(" ", "+")
    url = "https://www.kinopoisk.ru/index.php?kp_query=" + query

    response = session.get(url)
    response.raise_for_status()

    soup = bs(response.text, "html.parser")

    div_info = soup.find_all("div",class_="info")
    
    film_href = div_info[0].find_all("a", class_="js-serp-metrika")

    # Выцепляем ссылку на фильм
    film_path = ""
    for path in film_href:
        temp = path.get("href")
        if "film" in temp:
            film_path = temp
            break

    # Обрабатываем ошибку, если фильм не найден
    if film_path == "":
        print("Ошибка, фильм не найден")
        return
    
    # Формирование финальной ссылки на отзывы
    film_url = "https://www.kinopoisk.ru"+re.sub(r'(/film/\d+).*', r'\1', film_path)+"/reviews/ord/date/status/all/perpage/200/page/"

    print(film_url)

    time.sleep(1)
    # Цепляем отзывы
    response = session.get(film_url)
    response.raise_for_status()

    soup = bs(response.text, "html.parser")

    page = 1
    result = []
    while True:
        # Сам парсинг
        temp_result = soup.find_all("div", class_="brand_words")
        for review in temp_result:
            review_text = " ".join(review.text.split())
            result.append(review_text)

        # Логика прокрутки страниц
        time.sleep(5)

        div_counter = soup.select('div.pagesFromTo')
        if not div_counter:
            print("Пагинации нет → заканчиваем парсинг страницы.")
            break

        temp_list = div_counter[0].text.split(" из ")
        
        if int(temp_list[0].split("—")[1]) < int(temp_list[1]):
            page += 1
            response = session.get(film_url + str(page) + "/")
            response.raise_for_status()
            soup = bs(response.text, "html.parser")
        else:
            break

    if len(result) == 0:
        print("Нет отзывов")
        return
    return result

# https://www.kinopoisk.ru/index.php?kp_query=побег+из+шоушенка

# https://www.kinopoisk.ru/film/942396/reviews/ord/date/status/all/perpage/10/page/1/

from playwright.sync_api import sync_playwright
import re
import time

def parse_kp_info(raw_query):
    query = raw_query.strip().replace(" ", "+")
    search_url = f"https://www.kinopoisk.ru/index.php?kp_query={query}"

    with sync_playwright() as pw:
        browser = pw.chromium.launch(headless=False)  # True → без окна браузера
        page = browser.new_page()

        # 1. Переходим на страницу поиска
        page.goto(search_url, timeout=60000)
        page.wait_for_load_state("networkidle")

        # 2. Ищем ссылку на фильм
        film_link = page.locator("a.js-serp-metrika[href*='/film/']").first

        if not film_link:
            print("Фильм не найден")
            browser.close()
            return

        href = film_link.get_attribute("href")
        film_url = "https://www.kinopoisk.ru" + re.sub(r"(/film/\d+).*", r"\1", href)

        print("Ссылка на фильм:", film_url)

        # 3. Открываем страницу фильма
        page.goto(film_url, timeout=60000)
        page.wait_for_load_state("networkidle")

        # 4. Получаем заголовок
        title_el = page.locator("h1.styles_title__h1USJ").first
        
        if title_el.count() == 0:
            print("Не удалось получить название — возможно, страница под защитой")
        else:
            title = title_el.inner_text()
            print("Название:", title)

        browser.close()


# ТЕСТ
# parse_kp_info("побег из шоушенка")
parse_kp_reviews("побег из шоушенка")
