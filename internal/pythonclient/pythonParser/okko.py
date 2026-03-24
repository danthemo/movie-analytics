import os
import re
import sys
from urllib.parse import quote

from playwright.sync_api import TimeoutError as PlaywrightTimeoutError
from playwright.sync_api import sync_playwright


DEFAULT_TIMEOUT_MS = 30000
ANTI_BOT_MARKERS = ("xpvnsulc", "captcha", "challenge")


def clean_text(text):
    if not text:
        return ""
    return re.sub(r"\s+", " ", text).strip()


def is_debug_enabled():
    return os.getenv("OKKO_DEBUG", "").lower() in {"1", "true", "yes", "on"}


def launch_browser(browser_type):
    headless = os.getenv("OKKO_HEADLESS", "true").lower() not in {"0", "false", "no", "off"}
    slow_mo = 0
    if not headless:
        slow_mo = 500
    return browser_type.launch(headless=headless, slow_mo=slow_mo)


def log_debug(message):
    if is_debug_enabled():
        print(f"[OKKO DEBUG] {message}")


def wait_for_page(page):
    try:
        page.wait_for_load_state("domcontentloaded", timeout=DEFAULT_TIMEOUT_MS)
    except PlaywrightTimeoutError:
        log_debug("domcontentloaded timeout")

    try:
        page.wait_for_load_state("networkidle", timeout=5000)
    except PlaywrightTimeoutError:
        log_debug("networkidle timeout")

    page.wait_for_timeout(1500)


def ensure_not_blocked(page):
    current_url = page.url.lower()
    if any(marker in current_url for marker in ANTI_BOT_MARKERS):
        raise RuntimeError(f"okko anti-bot page detected: {page.url}")


def find_first_text(page, selectors):
    for selector in selectors:
        locator = page.locator(selector).first
        if locator.count() == 0:
            continue
        try:
            text = clean_text(locator.text_content())
        except Exception:
            continue
        if text:
            log_debug(f"text found for selector '{selector}': {text[:120]}")
            return text
    return ""


def find_list_text(page, selectors):
    for selector in selectors:
        locator = page.locator(selector)
        count = locator.count()
        if count == 0:
            continue
        values = []
        for idx in range(count):
            try:
                text = clean_text(locator.nth(idx).text_content())
            except Exception:
                continue
            if text:
                values.append(text)
        if values:
            log_debug(f"list found for selector '{selector}': {values[:5]}")
            return values
    return []


def get_candidate_links(page):
    selectors = [
        'a[test-id="search_prediction_unavailable_link"]',
        'a[test-id="search_collection_element"]',
        'div.Oce5ukhe a.vj8iwpkR',
        'a.vj8iwpkR',
        'a[href^="/movie/"]',
        'a[href*="/movie/"]',
        'a[href^="/serial/"]',
        'a[href*="/serial/"]',
    ]

    seen = set()
    candidates = []
    for selector in selectors:
        locator = page.locator(selector)
        count = locator.count()
        for idx in range(count):
            element = locator.nth(idx)
            href = element.get_attribute("href")
            text = clean_text(element.text_content())
            test_id = element.get_attribute("test-id")
            if not href or href in seen:
                continue
            seen.add(href)
            candidates.append({"href": href, "text": text, "test_id": test_id or ""})
    return candidates


def is_probable_title_match(query, text):
    normalized_query = clean_text(query).lower()
    normalized_text = clean_text(text).lower()
    if not normalized_query or not normalized_text:
        return False
    if normalized_query == normalized_text:
        return True
    return normalized_query in normalized_text or normalized_text in normalized_query


def exact_match_score(query, text):
    normalized_query = clean_text(query).lower()
    normalized_text = clean_text(text).lower()
    if not normalized_query or not normalized_text:
        return 0
    if normalized_query == normalized_text:
        return 3
    if normalized_text.startswith(normalized_query):
        return 2
    if normalized_query in normalized_text:
        return 1
    return 0


def find_film_link(page, query):
    candidates = get_candidate_links(page)
    log_debug(f"candidate links count: {len(candidates)}")

    best_match = None
    best_score = 0
    for candidate in candidates:
        score = exact_match_score(query, candidate["text"])
        if score > best_score:
            best_match = candidate["href"]
            best_score = score
        if score == 3:
            return candidate["href"]

    if best_match:
        return best_match

    for candidate in candidates:
        href = candidate["href"]
        if "/movie/" in href or "/serial/" in href:
            return href

    return None


def find_film_link_with_score(page, query):
    candidates = get_candidate_links(page)
    log_debug(f"candidate links count: {len(candidates)}")

    best_match = None
    best_score = 0
    for candidate in candidates:
        score = exact_match_score(query, candidate["text"])
        href = candidate["href"]

        if candidate.get("test_id") == "search_prediction_unavailable_link" and score > 0:
            return href, 10

        if score > best_score:
            best_match = href
            best_score = score
        if score == 3:
            return href, score

    if best_match:
        return best_match, best_score

    for candidate in candidates:
        href = candidate["href"]
        if "/movie/" in href or "/serial/" in href:
            return href, 0

    return None, 0


def ensure_absolute_okko_url(href):
    if href.startswith("http://") or href.startswith("https://"):
        return href
    return f"https://okko.tv{href}"


def is_film_page(page):
    return page.locator("h1").count() > 0


def extract_poster_url(page):
    selectors = [
        "picture source",
        "picture img",
        'img[test-id="poster"]',
        'img[alt*="постер"]',
        "img",
    ]

    for selector in selectors:
        locator = page.locator(selector).first
        if locator.count() == 0:
            continue

        if selector == "picture source":
            srcset = locator.get_attribute("srcset")
            if srcset:
                return srcset.split(" ")[0]
            continue

        for attr in ("src", "data-src"):
            value = locator.get_attribute(attr)
            if value and value.startswith("http"):
                return value

    return ""


def extract_year(page):
    year_text = find_first_text(
        page,
        [
            'span[test-id="meta_release_date"]',
            '[test-id="meta_release_date"]',
            '[data-testid="release-year"]',
        ],
    )
    match = re.search(r"(19|20)\d{2}", year_text)
    if match:
        return match.group(0)

    body_text = clean_text(page.text_content("body"))
    match = re.search(r"(19|20)\d{2}", body_text)
    if match:
        return match.group(0)
    return ""


def extract_description(page):
    return find_first_text(
        page,
        [
            'span[test-id="content_description"]',
            '[test-id="content_description"]',
            '[data-testid="content_description"]',
            "meta[name='description']",
        ],
    )


def extract_title(page):
    title = find_first_text(
        page,
        [
            "h1",
            '[data-testid="title"]',
            "title",
        ],
    )
    title = title.replace("Смотреть онлайн", "").replace("смотреть онлайн", "")
    return clean_text(title)


def parse_people_block(page, selectors):
    values = find_list_text(page, selectors)
    unique = []
    seen = set()
    for value in values:
        if value not in seen:
            seen.add(value)
            unique.append(value)
    return unique


def open_search_or_detect_direct_film(page, query):
    return search_via_input(page, query)


def search_via_input(page, query):
    page.goto("https://okko.tv", timeout=DEFAULT_TIMEOUT_MS)
    wait_for_page(page)
    ensure_not_blocked(page)

    search_button = page.locator('a[test-id="nav_search"]').first
    if search_button.count() == 0:
        log_debug("search button not found on okko main page")
        return None

    search_button.click()

    input_selectors = ['input[test-id="nav_search_input"]', 'input[type="search"]']
    search_input = None
    for selector in input_selectors:
        locator = page.locator(selector).first
        if locator.count() == 0:
            continue
        try:
            locator.wait_for(state="visible", timeout=5000)
        except PlaywrightTimeoutError:
            try:
                locator.wait_for(state="attached", timeout=5000)
            except PlaywrightTimeoutError:
                continue
        search_input = locator
        break

    if search_input is None:
        log_debug("search input not found after clicking search button")
        return None

    search_input.click()
    search_input.fill(query)
    page.wait_for_timeout(2000)

    try:
        page.wait_for_selector(
            'a[test-id="search_prediction_unavailable_link"], a[test-id="search_collection_element"]',
            state="attached",
            timeout=5000,
        )
    except PlaywrightTimeoutError:
        log_debug("no search result links attached after typing into search input")

    film_link, _ = find_film_link_with_score(page, query)
    return film_link


def parse_okko_film_info(query):
    with sync_playwright() as p:
        browser = launch_browser(p.chromium)
        try:
            page = browser.new_page()
            film_link = open_search_or_detect_direct_film(page, query)

            if film_link:
                page.goto(ensure_absolute_okko_url(film_link), timeout=DEFAULT_TIMEOUT_MS)
                wait_for_page(page)
                ensure_not_blocked(page)
            elif not is_film_page(page):
                return {}

            try:
                page.wait_for_selector("h1", state="attached", timeout=10000)
            except PlaywrightTimeoutError:
                log_debug(f"h1 was not attached on page: {page.url}")

            film_info = {
                "title": extract_title(page),
                "description": extract_description(page),
                "year": extract_year(page),
                "directors": parse_people_block(
                    page,
                    [
                        'div[test-id="meta_director"] a',
                        '[test-id="meta_director"] a',
                    ],
                ),
                "actors": parse_people_block(
                    page,
                    [
                        'div[test-id="meta_actors"] a',
                        '[test-id="meta_actors"] a',
                    ],
                ),
            }

            poster_url = extract_poster_url(page)
            if poster_url:
                film_info["poster_url"] = poster_url

            if not film_info["title"]:
                raise RuntimeError(f"failed to extract movie title from okko page: {page.url}")

            return film_info
        finally:
            browser.close()


def parse_okko_reviews(query):
    with sync_playwright() as p:
        browser = launch_browser(p.chromium)
        try:
            page = browser.new_page()
            film_link = open_search_or_detect_direct_film(page, query)

            if film_link:
                reviews_url = ensure_absolute_okko_url(film_link.rstrip("/") + "/reviews")
            elif is_film_page(page):
                reviews_url = page.url.rstrip("/") + "/reviews"
            else:
                return []

            page.goto(reviews_url, timeout=DEFAULT_TIMEOUT_MS)
            wait_for_page(page)
            ensure_not_blocked(page)

            candidate_selectors = [
                "div.GLR27TKM",
                '[data-testid="review-card"]',
                "article",
                "[class*='review']",
                "[class*='Review']",
            ]

            reviews = []
            seen = set()
            for selector in candidate_selectors:
                locator = page.locator(selector)
                count = locator.count()
                if count == 0:
                    continue
                for idx in range(count):
                    text = clean_text(locator.nth(idx).text_content())
                    if len(text) < 40:
                        continue
                    if text in seen:
                        continue
                    seen.add(text)
                    reviews.append(text)
                if reviews:
                    break

            return reviews
        finally:
            browser.close()


if __name__ == "__main__":
    query = " ".join(sys.argv[1:]).strip() or "Пчеловод"
    print("INFO:")
    print(parse_okko_film_info(query))
    print()
    print("REVIEWS:")
    print(parse_okko_reviews(query)[:3])
