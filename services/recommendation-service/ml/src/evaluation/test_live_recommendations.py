import os
from pathlib import Path

import joblib
import pandas as pd
import psycopg
from rerank import rerank_with_diversity

BASE_DIR = Path(__file__).resolve().parents[2]

RAW_PATH = (
    BASE_DIR
    / "data"
    / "raw"
    / "interactions.csv"
)

MODEL_PATH = (
    BASE_DIR
    / "models"
    / "logistic_regression.joblib"
)


NUMERIC_FEATURES = [
    "duration_seconds",
    "history_category_count",
    "history_category_positive",
    "history_category_negative",
    "history_category_likes",
    "history_category_unlikes",
    "history_category_skips",
    "history_category_positive_rate",
    "history_category_watch_ratio",
]

CATEGORICAL_FEATURES = [
    "categories",
]


CANDIDATE_QUERY = """
SELECT
    v.id AS video_id,
    v.duration_seconds,
    STRING_AGG(
        c.name,
        ',' ORDER BY c.name
    ) AS categories
FROM content_service.videos v
JOIN content_service.video_categories vc
    ON vc.video_id = v.id
JOIN content_service.categories c
    ON c.id = vc.category_id
WHERE v.status = 'READY'
GROUP BY
    v.id,
    v.duration_seconds
ORDER BY v.created_at DESC
LIMIT %s;
"""


def parse_categories(value) -> list[str]:
    if pd.isna(value):
        return []

    return [
        category.strip()
        for category in str(value).split(",")
        if category.strip()
    ]


def get_user_history(
    raw_df: pd.DataFrame,
    user_id: str,
) -> dict[str, dict]:
    user_df = raw_df[
        raw_df["user_id"] == user_id
    ].copy()

    if user_df.empty:
        raise ValueError(
            f"user not found: {user_id}"
        )

    history = {}

    for _, row in user_df.iterrows():
        categories = parse_categories(
            row["categories"]
        )

        watch_ratio = min(
            float(row["watched_seconds"])
            / float(row["duration_seconds"]),
            1.0,
        )

        positive = (
            bool(row["liked"])
            or watch_ratio >= 0.8
        )

        negative = (
            bool(row["unliked"])
            or bool(row["skipped"])
            or watch_ratio < 0.2
        )

        for category in categories:
            if category not in history:
                history[category] = {
                    "history_category_count": 0,
                    "history_category_positive": 0,
                    "history_category_negative": 0,
                    "history_category_likes": 0,
                    "history_category_unlikes": 0,
                    "history_category_skips": 0,
                    "watch_ratio_sum": 0.0,
                }

            stats = history[category]

            stats["history_category_count"] += 1
            stats["watch_ratio_sum"] += watch_ratio

            if row["liked"]:
                stats["history_category_likes"] += 1

            if row["unliked"]:
                stats["history_category_unlikes"] += 1

            if row["skipped"]:
                stats["history_category_skips"] += 1

            if positive:
                stats["history_category_positive"] += 1

            elif negative:
                stats["history_category_negative"] += 1

    for stats in history.values():
        count = stats["history_category_count"]

        if count > 0:
            stats["history_category_positive_rate"] = (
                stats["history_category_positive"]
                / count
            )

            stats["history_category_watch_ratio"] = (
                stats["watch_ratio_sum"]
                / count
            )

        else:
            stats["history_category_positive_rate"] = 0.5
            stats["history_category_watch_ratio"] = 0.5

        del stats["watch_ratio_sum"]

    return history


def load_live_candidates(
    database_url: str,
    limit: int,
) -> pd.DataFrame:
    with psycopg.connect(
        database_url
    ) as connection:

        with connection.cursor() as cursor:
            cursor.execute(
                CANDIDATE_QUERY,
                (limit,),
            )

            rows = cursor.fetchall()

            columns = [
                description.name
                for description
                in cursor.description
            ]

    return pd.DataFrame(
        rows,
        columns=columns,
    )


def remove_seen_videos(
    candidates: pd.DataFrame,
    raw_df: pd.DataFrame,
    user_id: str,
) -> pd.DataFrame:
    seen_videos = set(
        raw_df[
            raw_df["user_id"] == user_id
        ]["video_id"]
        .astype(str)
    )

    return candidates[
        ~candidates["video_id"]
        .astype(str)
        .isin(seen_videos)
    ].copy()


def build_candidate_features(
    candidates: pd.DataFrame,
    user_history: dict[str, dict],
) -> pd.DataFrame:
    rows = []

    for _, candidate in candidates.iterrows():
        categories = parse_categories(
            candidate["categories"]
        )

        category_stats = []

        for category in categories:
            if category in user_history:
                category_stats.append(
                    user_history[category]
                )

        if category_stats:
            history_count = sum(
                item["history_category_count"]
                for item in category_stats
            )

            history_positive = sum(
                item["history_category_positive"]
                for item in category_stats
            )

            history_negative = sum(
                item["history_category_negative"]
                for item in category_stats
            )

            history_likes = sum(
                item["history_category_likes"]
                for item in category_stats
            )

            history_unlikes = sum(
                item["history_category_unlikes"]
                for item in category_stats
            )

            history_skips = sum(
                item["history_category_skips"]
                for item in category_stats
            )

            weighted_watch_sum = sum(
                item["history_category_watch_ratio"]
                * item["history_category_count"]
                for item in category_stats
            )

            if history_count > 0:
                positive_rate = (
                    history_positive
                    / history_count
                )

                watch_ratio = (
                    weighted_watch_sum
                    / history_count
                )

            else:
                positive_rate = 0.5
                watch_ratio = 0.5

            stats = {
                "history_category_count":
                    history_count,

                "history_category_positive":
                    history_positive,

                "history_category_negative":
                    history_negative,

                "history_category_likes":
                    history_likes,

                "history_category_unlikes":
                    history_unlikes,

                "history_category_skips":
                    history_skips,

                "history_category_positive_rate":
                    positive_rate,

                "history_category_watch_ratio":
                    watch_ratio,
            }

        else:
            stats = {
                "history_category_count": 0,
                "history_category_positive": 0,
                "history_category_negative": 0,
                "history_category_likes": 0,
                "history_category_unlikes": 0,
                "history_category_skips": 0,
                "history_category_positive_rate": 0.5,
                "history_category_watch_ratio": 0.5,
            }

        rows.append(
            {
                "video_id":
                    candidate["video_id"],

                "categories":
                    candidate["categories"],

                "duration_seconds":
                    candidate["duration_seconds"],

                **stats,
            }
        )

    return pd.DataFrame(rows)


def main():
    database_url = os.getenv(
        "DATABASE_URL"
    )

    if not database_url:
        raise RuntimeError(
            "DATABASE_URL environment variable is required"
        )

    raw_df = pd.read_csv(
        RAW_PATH
    )

    raw_df.columns = (
        raw_df.columns
        .str.strip()
    )

    model = joblib.load(
        MODEL_PATH
    )

    user_counts = (
        raw_df["user_id"]
        .value_counts()
    )

    user_id = (
        user_counts.index[0]
    )

    print(
        f"Testing LIVE recommendations for user: {user_id}"
    )

    print(
        f"Known interactions: {user_counts.iloc[0]}"
    )

    user_history = get_user_history(
        raw_df,
        user_id,
    )

    print()
    print("Known category history:")

    for category, stats in sorted(
        user_history.items()
    ):
        print(
            f"{category:15} "
            f"count={stats['history_category_count']:5.1f} "
            f"positive_rate="
            f"{stats['history_category_positive_rate']:.2f} "
            f"watch_ratio="
            f"{stats['history_category_watch_ratio']:.2f}"
        )

    candidates = load_live_candidates(
        database_url,
        limit=2000,
    )

    print()
    print(
        f"Loaded live candidates: {len(candidates)}"
    )

    candidates = remove_seen_videos(
        candidates,
        raw_df,
        user_id,
    )

    print(
        f"Candidates after removing seen: {len(candidates)}"
    )

    candidate_features = (
        build_candidate_features(
            candidates,
            user_history,
        )
    )

    X = candidate_features[
        NUMERIC_FEATURES
        + CATEGORICAL_FEATURES
    ]

    probabilities = (
        model.predict_proba(
            X
        )[:, 1]
    )

    candidate_features[
        "score"
    ] = probabilities

    ranked_candidates = (
    candidate_features
    .sort_values(
        "score",
        ascending=False,
    )
)

    recommendations = rerank_with_diversity(
    ranked_candidates,
    limit=30,
    max_consecutive_same_category=2,
)

    print(
        "Top LIVE recommendations:"
    )

    print(
        recommendations[
            [
                "video_id",
                "categories",
                "duration_seconds",
                "score",
            ]
        ].to_string(
            index=False
        )
    )


if __name__ == "__main__":
    main()