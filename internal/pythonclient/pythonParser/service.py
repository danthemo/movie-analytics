from fastapi import FastAPI
from okko import parse_okko_film_info, parse_okko_reviews

app = FastAPI()

@app.get("/parse-reviews")
def get_reviews(query: str):
    reviews = parse_okko_reviews(query)
    return {"reviews": reviews}

@app.get("/parse-info")
def get_info(query: str):
    info = parse_okko_film_info(query)
    return {"info": info}
