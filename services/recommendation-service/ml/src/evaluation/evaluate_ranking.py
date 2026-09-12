from pathlib import Path
import math

import joblib
import numpy as np
import pandas as pd


BASE_DIR = Path(__file__).resolve().parents[2]

DATA_PATH = (
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


def evaluate_user(
    user_df: pd.DataFrame,
    model,
    k: int,
):
    X = user_df[
        NUMERIC_FEATURES
        + CATEGORICAL_FEATURES
    ]

    probabilities = (
        model.predict_proba(
            X
        )[:, 1]
    )

    scored = user_df.copy()

    scored["score"] = (
        probabilities
    )

    scored = scored.sort_values(
        "score",
        ascending=False,
    )

    relevance = (
        scored["target"]
        .to_numpy()
        .astype(int)
    )

    return {
        "precision": precision_at_k(
            relevance,
            k,
        ),
        "recall": recall_at_k(
            relevance,
            k,
        ),
        "hit_rate": hit_rate_at_k(
            relevance,
            k,
        ),
        "ndcg": ndcg_at_k(
            relevance,
            k,
        ),
    }


def main():
    df = pd.read_csv(
        DATA_PATH
    )

    df["started_at"] = (
        pd.to_datetime(
            df["started_at"],
            format="mixed",
            utc=True,
        )
    )

    df = df.sort_values(
        "started_at"
    ).reset_index(
        drop=True
    )

    split_index = int(
        len(df) * 0.8
    )

    test_df = (
        df.iloc[
            split_index:
        ]
        .copy()
    )

    model = joblib.load(
        MODEL_PATH
    )

    print(
        f"Test rows: {len(test_df)}"
    )

    print(
        f"Test users: "
        f"{test_df['user_id'].nunique()}"
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

        evaluated_users = 0

        for _, user_df in (
            test_df.groupby(
                "user_id"
            )
        ):
            if len(
                user_df
            ) < 2:
                continue

            if (
                user_df[
                    "target"
                ].sum()
                == 0
            ):
                continue

            metrics = (
                evaluate_user(
                    user_df,
                    model,
                    k,
                )
            )

            precision_scores.append(
                metrics[
                    "precision"
                ]
            )

            recall_scores.append(
                metrics[
                    "recall"
                ]
            )

            hit_rate_scores.append(
                metrics[
                    "hit_rate"
                ]
            )

            ndcg_scores.append(
                metrics[
                    "ndcg"
                ]
            )

            evaluated_users += 1

        print()
        print(
            f"Ranking metrics @ {k}"
        )

        print(
            f"Evaluated users: "
            f"{evaluated_users}"
        )

        print(
            "Precision@"
            f"{k}: "
            f"{np.mean(precision_scores):.4f}"
        )

        print(
            "Recall@"
            f"{k}:    "
            f"{np.mean(recall_scores):.4f}"
        )

        print(
            "HitRate@"
            f"{k}:   "
            f"{np.mean(hit_rate_scores):.4f}"
        )

        print(
            "NDCG@"
            f"{k}:      "
            f"{np.mean(ndcg_scores):.4f}"
        )


if __name__ == "__main__":
    main()