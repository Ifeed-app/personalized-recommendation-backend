from collections import defaultdict
from pathlib import Path

import numpy as np
import pandas as pd

RAW_PATH = (
    Path(__file__).resolve().parents[2]
    / "data"
    / "raw"
    / "interactions.csv"
)

OUTPUT_PATH = (
    Path(__file__).resolve().parents[2]
    / "data"
    / "processed"
    / "features.csv"
)


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

    return np.nan


def parse_categories(value: str) -> list[str]:
    if pd.isna(value):
        return []

    return [
        category.strip()
        for category in value.split(",")
        if category.strip()
    ]


def main():
    df = pd.read_csv(RAW_PATH)

    df["started_at"] = pd.to_datetime(
    df["started_at"],
    format="mixed",
    utc=True,
)

    df = df.sort_values(
        ["user_id", "started_at"]
    ).reset_index(drop=True)

    df["watch_ratio"] = (
        df["watched_seconds"]
        / df["duration_seconds"]
    ).clip(0, 1)

    user_history = defaultdict(
        lambda: defaultdict(
            lambda: {
                "count": 0,
                "positive": 0,
                "negative": 0,
                "likes": 0,
                "unlikes": 0,
                "skips": 0,
                "watch_ratio_sum": 0.0,
            }
        )
    )

    output_rows = []

    for _, row in df.iterrows():
        user_id = row["user_id"]

        categories = parse_categories(
            row["categories"]
        )

        watch_ratio = float(
            row["watch_ratio"]
        )

        target = build_target(
            liked=bool(row["liked"]),
            unliked=bool(row["unliked"]),
            skipped=bool(row["skipped"]),
            watch_ratio=watch_ratio,
        )

        history_count = 0
        history_positive = 0
        history_negative = 0

        history_likes = 0
        history_unlikes = 0
        history_skips = 0

        history_watch_ratio_sum = 0.0

        for category in categories:
            stats = user_history[user_id][category]

            history_count += stats["count"]
            history_positive += stats["positive"]
            history_negative += stats["negative"]

            history_likes += stats["likes"]
            history_unlikes += stats["unlikes"]
            history_skips += stats["skips"]

            history_watch_ratio_sum += stats[
                "watch_ratio_sum"
            ]

        if history_count > 0:
            historical_positive_rate = (
                history_positive
                / history_count
            )

            historical_watch_ratio = (
                history_watch_ratio_sum
                / history_count
            )
        else:
            historical_positive_rate = 0.5
            historical_watch_ratio = 0.5

        #
        # Samo označeni redovi ulaze u training dataset.
        #
        if (
            categories
            and not pd.isna(target)
        ):
            output_rows.append(
                {
                    "user_id": user_id,
                    "video_id": row["video_id"],
                    "started_at": row["started_at"],

                    "categories": ",".join(categories),

                    "duration_seconds": row[
                        "duration_seconds"
                    ],

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
                        historical_positive_rate,

                    "history_category_watch_ratio":
                        historical_watch_ratio,

                    "target": int(target),
                }
            )

        for category in categories:
            stats = user_history[
                user_id
            ][category]

            stats["count"] += 1

            stats["watch_ratio_sum"] += (
                watch_ratio
            )

            if row["liked"]:
                stats["likes"] += 1

            if row["unliked"]:
                stats["unlikes"] += 1

            if row["skipped"]:
                stats["skips"] += 1

            if not pd.isna(target):
                if target == 1:
                    stats["positive"] += 1
                else:
                    stats["negative"] += 1

    features = pd.DataFrame(
        output_rows
    )

    OUTPUT_PATH.parent.mkdir(
        parents=True,
        exist_ok=True,
    )

    features.to_csv(
        OUTPUT_PATH,
        index=False,
    )

    print("Feature dataset created.")
    print(f"Rows: {len(features)}")
    print(
        f"Users: {features['user_id'].nunique()}"
    )
    print(
        f"Videos: {features['video_id'].nunique()}"
    )

    print()
    print("Target distribution:")
    print(
        features["target"].value_counts()
    )

    print()
    print("History count:")
    print(
        features[
            "history_category_count"
        ].describe()
    )

    print()
    print(
        "Rows with category history:",
        (
            features[
                "history_category_count"
            ] > 0
        ).sum(),
    )


if __name__ == "__main__":
    main()