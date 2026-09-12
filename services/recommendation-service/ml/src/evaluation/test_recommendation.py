from pathlib import Path

import joblib
import pandas as pd


BASE_DIR = Path(__file__).resolve().parents[2]

RAW_PATH = (
    BASE_DIR
    / "data"
    / "raw"
    / "interactions.csv"
)

FEATURES_PATH = (
    BASE_DIR
    / "data"
    / "processed"
    / "features.csv"
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
    required_columns = {
        "user_id",
        "categories",
        "watched_seconds",
        "duration_seconds",
        "liked",
        "unliked",
        "skipped",
    }

    missing_columns = (
        required_columns
        - set(raw_df.columns)
    )

    if missing_columns:
        raise ValueError(
            f"Missing raw columns: {missing_columns}"
        )

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
                stats["watch_ratio_sum"]
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


def get_candidate_videos(
    df: pd.DataFrame,
    user_id: str,
    limit: int = 100,
) -> pd.DataFrame:
    seen_videos = set(
        df[
            df["user_id"] == user_id
        ]["video_id"]
    )

    candidates = df[
        ~df["video_id"].isin(
            seen_videos
        )
    ][
        [
            "video_id",
            "categories",
            "duration_seconds",
        ]
    ].drop_duplicates(
        subset=["video_id"]
    )

    return (
        candidates
        .head(limit)
        .copy()
    )


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
                    candidate[
                        "duration_seconds"
                    ],

                **stats,
            }
        )

    return pd.DataFrame(rows)


def main():
    features_df = pd.read_csv(
        FEATURES_PATH
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
        features_df[
            "user_id"
        ]
        .value_counts()
    )

    user_id = (
        user_counts.index[0]
    )

    print(
        f"Testing recommendations for user: {user_id}"
    )

    print(
        f"Known interactions: {user_counts.iloc[0]}"
    )

    user_history = (
        get_user_history(
            raw_df,
            user_id,
        )
    )

    print()
    print(
        "Known category history:"
    )

    for category, stats in sorted(
        user_history.items()
    ):
        print(
            f"{category:15} "
            f"count="
            f"{stats['history_category_count']:5.1f} "
            f"positive_rate="
            f"{stats['history_category_positive_rate']:.2f} "
            f"watch_ratio="
            f"{stats['history_category_watch_ratio']:.2f}"
        )

    candidates = (
        get_candidate_videos(
            features_df,
            user_id,
            limit=100,
        )
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

    recommendations = (
        candidate_features
        .sort_values(
            "score",
            ascending=False,
        )
        .head(20)
    )

    print()
    print(
        "Top recommendations:"
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