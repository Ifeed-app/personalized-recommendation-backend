from pathlib import Path

import joblib
import pandas as pd

MODEL_PATH = (
    Path(__file__).resolve().parents[2]
    / "models"
    / "logistic_regression.joblib"
)


def main():
    pipeline = joblib.load(MODEL_PATH)

    preprocessor = pipeline.named_steps["preprocessor"]
    model = pipeline.named_steps["model"]

    feature_names = preprocessor.get_feature_names_out()

    coefficients = model.coef_[0]

    result = pd.DataFrame(
        {
            "feature": feature_names,
            "coefficient": coefficients,
            "abs_coefficient": abs(coefficients),
        }
    )

    result = result.sort_values(
        "abs_coefficient",
        ascending=False,
    )

    print(result.head(30).to_string(index=False))


if __name__ == "__main__":
    main()