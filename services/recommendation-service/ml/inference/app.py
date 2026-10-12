from pathlib import Path
from typing import List

import joblib
import pandas as pd
from fastapi import FastAPI
from pydantic import BaseModel


BASE_DIR = Path(__file__).resolve().parent.parent
MODEL_PATH = BASE_DIR / "models" / "logistic_regression.joblib"

model = joblib.load(MODEL_PATH)

app = FastAPI(
    title="Feeder ML Inference Service",
    version="1.0.0",
)


class Candidate(BaseModel):
    video_id: str
    duration_seconds: float

    history_category_count: float
    history_category_positive: float
    history_category_negative: float
    history_category_likes: float
    history_category_unlikes: float
    history_category_skips: float
    history_category_positive_rate: float
    history_category_watch_ratio: float

    categories: str


class ScoreRequest(BaseModel):
    candidates: List[Candidate]


class ScoredCandidate(BaseModel):
    video_id: str
    score: float


class ScoreResponse(BaseModel):
    scores: List[ScoredCandidate]


@app.get("/health")
def health():
    return {
        "status": "ok",
        "model": "logistic_regression",
    }


@app.post("/score", response_model=ScoreResponse)
def score(request: ScoreRequest):
    if not request.candidates:
        return ScoreResponse(scores=[])

    rows = []

    for candidate in request.candidates:
        rows.append({
            "duration_seconds": candidate.duration_seconds,
            "history_category_count": candidate.history_category_count,
            "history_category_positive": candidate.history_category_positive,
            "history_category_negative": candidate.history_category_negative,
            "history_category_likes": candidate.history_category_likes,
            "history_category_unlikes": candidate.history_category_unlikes,
            "history_category_skips": candidate.history_category_skips,
            "history_category_positive_rate": candidate.history_category_positive_rate,
            "history_category_watch_ratio": candidate.history_category_watch_ratio,
            "categories": candidate.categories,
        })

    df = pd.DataFrame(rows)

    raw_scores = model.decision_function(df)

    scores = [
        ScoredCandidate(
            video_id=c.video_id,
            score=float(score),
        )
        for c, score in zip(
            request.candidates,
            raw_scores,
        )
    ]

    return ScoreResponse(
        scores=scores,
    )