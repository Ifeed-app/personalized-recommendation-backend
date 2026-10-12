import joblib
import numpy as np
import matplotlib.pyplot as plt
import pandas as pd

MODEL_PATH = "models/hist_gradient_boosting.joblib"

model = joblib.load(MODEL_PATH)
print(model.named_steps.keys())
preprocessor = model.named_steps["preprocessor"]
classifier = model.named_steps["model"]
feature_names = preprocessor.get_feature_names_out()

bin_mapper = classifier._bin_mapper

for i, name in enumerate(feature_names):
    thresholds = bin_mapper.bin_thresholds_[i]

    print(f"\nFeature {i}: {name}")
    print(f"Broj granica: {len(thresholds)}")
    print(thresholds[:20])

target_feature = "numeric__history_category_watch_ratio"

feature_index = list(feature_names).index(target_feature)
thresholds = bin_mapper.bin_thresholds_[feature_index]

plt.figure(figsize=(10, 5))

plt.hist(
    thresholds,
    bins=30
)

plt.xlabel("Vrijednost granice bina")
plt.ylabel("Broj granica")
plt.title("Histogram granica binova za history_category_watch_ratio")

plt.tight_layout()
plt.savefig("hist_gradient_boosting_bins.png", dpi=300)
plt.show()
DATA_PATH = "data/processed/features.csv"

df = pd.read_csv(DATA_PATH)

feature_column = "history_category_watch_ratio"
target_feature = "numeric__history_category_watch_ratio"

feature_index = list(feature_names).index(target_feature)
thresholds = bin_mapper.bin_thresholds_[feature_index]

values = df[feature_column].dropna()

plt.figure(figsize=(10, 5))

# Histogram stvarnih vrijednosti značajke
plt.hist(
    values,
    bins=30,
    alpha=0.7,
    edgecolor="black"
)

# Nacrtaj dio stvarnih bin granica modela
# Svaku 10. granicu da graf ne bude prenatrpan
for threshold in thresholds[::10]:
    plt.axvline(
        threshold,
        linestyle="--",
        linewidth=1
    )

plt.xlabel("history_category_watch_ratio")
plt.ylabel("Broj zapisa")
plt.title(
    "Histogram značajke history_category_watch_ratio "
    "s granicama binova modela"
)

plt.tight_layout()

plt.savefig(
    "hist_gradient_boosting_watch_ratio_bins.png",
    dpi=300
)