from pathlib import Path

import joblib
import pandas as pd

from sklearn.compose import ColumnTransformer
from sklearn.linear_model import LogisticRegression
from sklearn.metrics import (
    accuracy_score,
    classification_report,
    confusion_matrix,
    roc_auc_score,
)
from sklearn.pipeline import Pipeline
from sklearn.preprocessing import OneHotEncoder, StandardScaler


DATA_PATH = (
    Path(__file__).resolve().parents[2]
    / "data"
    / "processed"
    / "features.csv"
)

MODEL_PATH = (
    Path(__file__).resolve().parents[2]
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


def main():
    df = pd.read_csv(DATA_PATH)

    df["started_at"] = pd.to_datetime(
        df["started_at"],
        format="mixed",
        utc=True,
    )

    #
    # Temporal split:
    # starijih 80% za trening,
    # najnovijih 20% za test.
    #
    df = df.sort_values(
        "started_at"
    ).reset_index(drop=True)

    split_index = int(len(df) * 0.8)

    train_df = df.iloc[:split_index]
    test_df = df.iloc[split_index:]

    X_train = train_df[
        NUMERIC_FEATURES + CATEGORICAL_FEATURES
    ]

    y_train = train_df["target"]

    X_test = test_df[
        NUMERIC_FEATURES + CATEGORICAL_FEATURES
    ]

    y_test = test_df["target"]

    numeric_transformer = Pipeline(
        steps=[
            (
                "scaler",
                StandardScaler(),
            ),
        ]
    )

    categorical_transformer = Pipeline(
        steps=[
            (
                "onehot",
                OneHotEncoder(
                    handle_unknown="ignore",
                ),
            ),
        ]
    )

    preprocessor = ColumnTransformer(
        transformers=[
            (
                "numeric",
                numeric_transformer,
                NUMERIC_FEATURES,
            ),
            (
                "categorical",
                categorical_transformer,
                CATEGORICAL_FEATURES,
            ),
        ]
    )

    model = LogisticRegression(
        max_iter=1000,
    )

    pipeline = Pipeline(
        steps=[
            (
                "preprocessor",
                preprocessor,
            ),
            (
                "model",
                model,
            ),
        ]
    )

    print(
        f"Train rows: {len(train_df)}"
    )
    print(
        f"Test rows: {len(test_df)}"
    )

    pipeline.fit(
        X_train,
        y_train,
    )

    predictions = pipeline.predict(
        X_test
    )

    probabilities = pipeline.predict_proba(
        X_test
    )[:, 1]

    accuracy = accuracy_score(
        y_test,
        predictions,
    )

    roc_auc = roc_auc_score(
        y_test,
        probabilities,
    )

    print()
    print(
        f"Accuracy: {accuracy:.4f}"
    )
    print(
        f"ROC AUC:  {roc_auc:.4f}"
    )

    print()
    print("Classification report:")
    print(
        classification_report(
            y_test,
            predictions,
        )
    )

    print("Confusion matrix:")
    print(
        confusion_matrix(
            y_test,
            predictions,
        )
    )

    MODEL_PATH.parent.mkdir(
        parents=True,
        exist_ok=True,
    )

    joblib.dump(
        pipeline,
        MODEL_PATH,
    )

    print()
    print(
        f"Model saved to: {MODEL_PATH}"
    )


if __name__ == "__main__":
    main()