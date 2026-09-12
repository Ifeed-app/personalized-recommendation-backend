import pandas as pd


def rerank_with_diversity(
    recommendations: pd.DataFrame,
    limit: int = 20,
    max_consecutive_same_category: int = 2,
) -> pd.DataFrame:
    remaining = recommendations.copy()

    selected_rows = []
    last_category = None
    consecutive_count = 0

    while (
        len(selected_rows) < limit
        and not remaining.empty
    ):
        selected_index = None

        for index, row in remaining.iterrows():
            category = row["categories"]

            if (
                category == last_category
                and consecutive_count
                >= max_consecutive_same_category
            ):
                continue

            selected_index = index
            break

        if selected_index is None:
            selected_index = remaining.index[0]

        row = remaining.loc[selected_index]

        category = row["categories"]

        if category == last_category:
            consecutive_count += 1
        else:
            last_category = category
            consecutive_count = 1

        selected_rows.append(
            row.to_dict()
        )

        remaining = remaining.drop(
            index=selected_index
        )

    return pd.DataFrame(
        selected_rows
    )