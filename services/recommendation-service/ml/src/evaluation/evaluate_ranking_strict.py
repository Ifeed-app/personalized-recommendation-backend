from pathlib import Path
import math

import joblib
import numpy as np
import pandas as pd
import sys


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


NEGATIVES_PER_POSITIVE = 20
RANDOM_SEED = 42
TRAIN_RATIO = 0.8


def parse_categories(value) -> list[str]:
    if pd.isna(value):
        return []

    return [
        category.strip()
        for category in str(value).split(",")
        if category.strip()
    ]


def build_target(
    liked: bool,
    unliked: bool,
    skipped: bool,
    watch_ratio: float,
):
    if liked:
        return 1

    if unliked or skipped:
        return 0

    if watch_ratio >= 0.8:
        return 1

    if watch_ratio < 0.2:
        return 0

    return None


def precision_at_k(
    relevance: np.ndarray,
    k: int,
) -> float:
    top_k = relevance[:k]

    if len(top_k) == 0:
        return 0.0

    return float(
        np.sum(top_k)
        / len(top_k)
    )


def recall_at_k(
    relevance: np.ndarray,
    k: int,
) -> float:
    total_relevant = np.sum(
        relevance
    )

    if total_relevant == 0:
        return 0.0

    return float(
        np.sum(relevance[:k])
        / total_relevant
    )


def hit_rate_at_k(
    relevance: np.ndarray,
    k: int,
) -> float:
    return float(
        np.any(
            relevance[:k] == 1
        )
    )


def dcg_at_k(
    relevance: np.ndarray,
    k: int,
) -> float:
    relevance = relevance[:k]

    score = 0.0

    for index, rel in enumerate(
        relevance
    ):
        rank = index + 1

        score += (
            rel
            / math.log2(
                rank + 1
            )
        )

    return score


def ndcg_at_k(
    relevance: np.ndarray,
    k: int,
) -> float:
    actual_dcg = dcg_at_k(
        relevance,
        k,
    )

    ideal_relevance = np.sort(
        relevance
    )[::-1]

    ideal_dcg = dcg_at_k(
        ideal_relevance,
        k,
    )

    if ideal_dcg == 0:
        return 0.0

    return actual_dcg / ideal_dcg


def build_frozen_user_history(
    history_df: pd.DataFrame,
) -> dict[str, dict]:
    history = {}

    for _, row in history_df.iterrows():
        categories = parse_categories(
            row["categories"]
        )

        watch_ratio = min(
            float(row["watched_seconds"])
            / float(row["duration_seconds"]),
            1.0,
        )

        target = build_target(
            liked=bool(row["liked"]),
            unliked=bool(row["unliked"]),
            skipped=bool(row["skipped"]),
            watch_ratio=watch_ratio,
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

            stats[
                "history_category_count"
            ] += 1

            stats[
                "watch_ratio_sum"
            ] += watch_ratio

            if row["liked"]:
                stats[
                    "history_category_likes"
                ] += 1

            if row["unliked"]:
                stats[
                    "history_category_unlikes"
                ] += 1

            if row["skipped"]:
                stats[
                    "history_category_skips"
                ] += 1

            if target == 1:
                stats[
                    "history_category_positive"
                ] += 1

            elif target == 0:
                stats[
                    "history_category_negative"
                ] += 1

    for stats in history.values():
        count = stats[
            "history_category_count"
        ]

        if count > 0:
            stats[
                "history_category_positive_rate"
            ] = (
                stats[
                    "history_category_positive"
                ]
                / count
            )

            stats[
                "history_category_watch_ratio"
            ] = (
                stats[
                    "watch_ratio_sum"
                ]
                / count
            )

        else:
            stats[
                "history_category_positive_rate"
            ] = 0.5

            stats[
                "history_category_watch_ratio"
            ] = 0.5

        del stats["watch_ratio_sum"]

    return history


def build_candidate_features(
    video_id,
    categories,
    duration_seconds,
    user_history,
    target,
) -> dict:
    parsed_categories = parse_categories(
        categories
    )

    category_stats = [
        user_history[category]
        for category in parsed_categories
        if category in user_history
    ]

    if category_stats:
        history_count = sum(
            item[
                "history_category_count"
            ]
            for item in category_stats
        )

        history_positive = sum(
            item[
                "history_category_positive"
            ]
            for item in category_stats
        )

        history_negative = sum(
            item[
                "history_category_negative"
            ]
            for item in category_stats
        )

        history_likes = sum(
            item[
                "history_category_likes"
            ]
            for item in category_stats
        )

        history_unlikes = sum(
            item[
                "history_category_unlikes"
            ]
            for item in category_stats
        )

        history_skips = sum(
            item[
                "history_category_skips"
            ]
            for item in category_stats
        )

        weighted_watch_sum = sum(
            item[
                "history_category_watch_ratio"
            ]
            * item[
                "history_category_count"
            ]
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

    else:
        history_count = 0
        history_positive = 0
        history_negative = 0
        history_likes = 0
        history_unlikes = 0
        history_skips = 0
        positive_rate = 0.5
        watch_ratio = 0.5

    return {
        "video_id":
            video_id,

        "categories":
            categories,

        "duration_seconds":
            duration_seconds,

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

        "target":
            target,
    }


def build_candidate_pool_for_user(
    user_df: pd.DataFrame,
    all_videos: pd.DataFrame,
    rng: np.random.Generator,
) -> pd.DataFrame:
    user_df = (
        user_df
        .sort_values("started_at")
        .reset_index(drop=True)
    )

    if len(user_df) < 5:
        return pd.DataFrame()

    split_index = int(
        len(user_df)
        * TRAIN_RATIO
    )

    history_df = (
        user_df.iloc[
            :split_index
        ]
        .copy()
    )

    holdout_df = (
        user_df.iloc[
            split_index:
        ]
        .copy()
    )

    user_history = (
        build_frozen_user_history(
            history_df
        )
    )

    positive_holdout_rows = []

    for _, row in holdout_df.iterrows():
        watch_ratio = min(
            float(row["watched_seconds"])
            / float(row["duration_seconds"]),
            1.0,
        )

        target = build_target(
            liked=bool(row["liked"]),
            unliked=bool(row["unliked"]),
            skipped=bool(row["skipped"]),
            watch_ratio=watch_ratio,
        )

        if target == 1:
            positive_holdout_rows.append(
                row
            )

    if not positive_holdout_rows:
        return pd.DataFrame()

    positive_video_ids = {
        row["video_id"]
        for row in positive_holdout_rows
    }

    all_seen_video_ids = set(
        user_df["video_id"]
    )

    negative_pool = (
        all_videos[
            ~all_videos["video_id"].isin(
                all_seen_video_ids
            )
        ]
        .drop_duplicates(
            subset=["video_id"]
        )
        .copy()
    )

    negative_count = (
        len(positive_holdout_rows)
        * NEGATIVES_PER_POSITIVE
    )

    negative_count = min(
        negative_count,
        len(negative_pool),
    )

    if negative_count == 0:
        return pd.DataFrame()

    sampled_indices = rng.choice(
        negative_pool.index.to_numpy(),
        size=negative_count,
        replace=False,
    )

    sampled_negatives = (
        negative_pool
        .loc[
            sampled_indices
        ]
        .copy()
    )

    candidate_rows = []

    for row in positive_holdout_rows:
        candidate_rows.append(
            build_candidate_features(
                video_id=row["video_id"],
                categories=row["categories"],
                duration_seconds=row[
                    "duration_seconds"
                ],
                user_history=user_history,
                target=1,
            )
        )

    for _, row in sampled_negatives.iterrows():
        if row["video_id"] in positive_video_ids:
            continue

        candidate_rows.append(
            build_candidate_features(
                video_id=row["video_id"],
                categories=row["categories"],
                duration_seconds=row[
                    "duration_seconds"
                ],
                user_history=user_history,
                target=0,
            )
        )

    return pd.DataFrame(
        candidate_rows
    )


def evaluate_user(
    candidate_df: pd.DataFrame,
    model,
    k: int,
) -> dict:
    X = candidate_df[
        NUMERIC_FEATURES
        + CATEGORICAL_FEATURES
    ]

    probabilities = (
        model.predict_proba(
            X
        )[:, 1]
    )

    scored = (
        candidate_df
        .copy()
    )

    scored["score"] = (
        probabilities
    )

    scored = (
        scored
        .sort_values(
            "score",
            ascending=False,
        )
        .reset_index(drop=True)
    )

    relevance = (
        scored[
            "target"
        ]
        .to_numpy()
        .astype(int)
    )

    return {
        "precision":
            precision_at_k(
                relevance,
                k,
            ),

        "recall":
            recall_at_k(
                relevance,
                k,
            ),

        "hit_rate":
            hit_rate_at_k(
                relevance,
                k,
            ),

        "ndcg":
            ndcg_at_k(
                relevance,
                k,
            ),
    }


def main():
    raw_df = pd.read_csv(
        RAW_PATH
    )

    raw_df.columns = (
        raw_df.columns
        .str.strip()
    )

    raw_df["started_at"] = (
        pd.to_datetime(
            raw_df["started_at"],
            format="mixed",
            utc=True,
        )
    )

    raw_df = (
        raw_df
        .sort_values(
            [
                "user_id",
                "started_at",
            ]
        )
        .reset_index(drop=True)
    )

    all_videos = (
        raw_df[
            [
                "video_id",
                "categories",
                "duration_seconds",
            ]
        ]
        .drop_duplicates(
            subset=["video_id"]
        )
        .copy()
    )

    if len(sys.argv) > 1:
        model_path = Path(sys.argv[1])
    else:
        model_path = (
            BASE_DIR
            / "models"
            / "logistic_regression.joblib"
        )

    print(
        f"Loading model: {model_path}"
    )

    model = joblib.load(
        model_path
    )
    rng = np.random.default_rng(
        RANDOM_SEED
    )

    print(
        f"Users in raw dataset: "
        f"{raw_df['user_id'].nunique()}"
    )

    print(
        f"Videos in candidate universe: "
        f"{len(all_videos)}"
    )

    print(
        f"Per-user history ratio: "
        f"{TRAIN_RATIO:.0%}"
    )

    print(
        f"Negative samples per positive: "
        f"{NEGATIVES_PER_POSITIVE}"
    )

    candidate_pools = {}

    for user_id, user_df in (
        raw_df.groupby(
            "user_id"
        )
    ):
        candidate_pool = (
            build_candidate_pool_for_user(
                user_df,
                all_videos,
                rng,
            )
        )

        if candidate_pool.empty:
            continue

        if (
            candidate_pool[
                "target"
            ].sum()
            == 0
        ):
            continue

        candidate_pools[
            user_id
        ] = candidate_pool

    print(
        f"Evaluated users available: "
        f"{len(candidate_pools)}"
    )

    for k in [
        5,
        10,
        20,
    ]:
        precision_scores = []
        recall_scores = []
        hit_rate_scores = []
        ndcg_scores = []

        for candidate_pool in (
            candidate_pools.values()
        ):
            metrics = evaluate_user(
                candidate_pool,
                model,
                k,
            )

            precision_scores.append(
                metrics["precision"]
            )

            recall_scores.append(
                metrics["recall"]
            )

            hit_rate_scores.append(
                metrics["hit_rate"]
            )

            ndcg_scores.append(
                metrics["ndcg"]
            )

        print()
        print(
            f"Strict frozen-history "
            f"ranking metrics @ {k}"
        )

        print(
            f"Evaluated users: "
            f"{len(candidate_pools)}"
        )

        print(
            f"Precision@{k}: "
            f"{np.mean(precision_scores):.4f}"
        )

        print(
            f"Recall@{k}:    "
            f"{np.mean(recall_scores):.4f}"
        )

        print(
            f"HitRate@{k}:   "
            f"{np.mean(hit_rate_scores):.4f}"
        )

        print(
            f"NDCG@{k}:      "
            f"{np.mean(ndcg_scores):.4f}"
        )


if __name__ == "__main__":
    main()