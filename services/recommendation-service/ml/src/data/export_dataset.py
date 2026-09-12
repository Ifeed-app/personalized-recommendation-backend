from pathlib import Path
import os

import pandas as pd
import psycopg


OUTPUT_PATH = Path(__file__).resolve().parents[2] / "data" / "raw" / "interactions.csv"


QUERY = """
WITH interaction_flags AS (
    SELECT
        user_id,
        video_id,
        BOOL_OR(interaction_type = 'LIKE') AS liked,
        BOOL_OR(interaction_type = 'UNLIKE') AS unliked,
        BOOL_OR(interaction_type = 'SKIP') AS skipped
    FROM integration_service.video_interactions
    GROUP BY user_id, video_id
),

video_category_data AS (
    SELECT
        vc.video_id,
        STRING_AGG(c.name, ',' ORDER BY c.name) AS categories
    FROM content_service.video_categories vc
    JOIN content_service.categories c
        ON c.id = vc.category_id
    GROUP BY vc.video_id
)

SELECT
    ws.id AS watch_session_id,
    ws.user_id,
    ws.video_id,

    ws.watched_seconds,
    ws.completed,
    ws.started_at,
    ws.ended_at,

    v.duration_seconds,

    COALESCE(i.liked, FALSE) AS liked,
    COALESCE(i.unliked, FALSE) AS unliked,
    COALESCE(i.skipped, FALSE) AS skipped,

    COALESCE(vc.categories, '') AS categories

FROM integration_service.video_watch_sessions ws

JOIN content_service.videos v
    ON v.id = ws.video_id

LEFT JOIN interaction_flags i
    ON i.user_id = ws.user_id
    AND i.video_id = ws.video_id

LEFT JOIN video_category_data vc
    ON vc.video_id = ws.video_id

WHERE v.status = 'READY'

ORDER BY ws.started_at;
"""


def main():
    database_url = os.getenv("DATABASE_URL")

    if not database_url:
        raise RuntimeError("DATABASE_URL environment variable is required")

    print("Connecting to PostgreSQL...")

    with psycopg.connect(database_url) as connection:
        dataframe = pd.read_sql_query(
            QUERY,
            connection,
        )

    OUTPUT_PATH.parent.mkdir(
        parents=True,
        exist_ok=True,
    )

    dataframe.to_csv(
        OUTPUT_PATH,
        index=False,
    )

    print()
    print("Dataset exported successfully.")
    print(f"Rows: {len(dataframe)}")
    print(f"Users: {dataframe['user_id'].nunique()}")
    print(f"Videos: {dataframe['video_id'].nunique()}")
    print(f"Output: {OUTPUT_PATH}")


if __name__ == "__main__":
    main()