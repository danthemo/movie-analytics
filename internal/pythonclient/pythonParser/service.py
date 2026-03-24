import logging

from fastapi import FastAPI, HTTPException
from okko import parse_okko_film_info, parse_okko_reviews

app = FastAPI()
logger = logging.getLogger(__name__)

@app.get("/parse-reviews")
def get_reviews(query: str):
    try:
        reviews = parse_okko_reviews(query)
        return {"reviews": reviews}
    except Exception as exc:
        logger.exception("Failed to parse reviews for query '%s'", query)
        raise HTTPException(status_code=502, detail="failed to parse reviews") from exc

@app.get("/parse-info")
def get_info(query: str):
    try:
        info = parse_okko_film_info(query)
        return {"info": info}
    except Exception as exc:
        logger.exception("Failed to parse film info for query '%s'", query)
        raise HTTPException(status_code=502, detail="failed to parse film info") from exc
