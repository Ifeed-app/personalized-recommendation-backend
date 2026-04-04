CREATE TABLE IF NOT EXISTS integration_service.video_interactions(
id UUID PRIMARY KEY,
user_id UUID NOT NULL,
video_id UUID NOT NULL,
interaction_type TEXT NOT NULL,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS integration_service.video_watch_sessions (
id UUID PRIMARY KEY,
user_id UUID NOT NULL,
video_id UUID NOT NULL,
watched_seconds INTEGER NOT NULL DEFAULT 0,
completed BOOLEAN NOT NULL DEFAULT FALSE,
started_at TIMESTAMPTZ NOT NULL,
ended_at TIMESTAMPTZ,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_video_interactions_user_id
    ON integration_service.video_interactions(user_id);

CREATE INDEX IF NOT EXISTS idx_video_interactions_video_id
    ON integration_service.video_interactions(video_id);

CREATE INDEX IF NOT EXISTS idx_video_interactions_type
    ON integration_service.video_interactions(interaction_type);

CREATE INDEX IF NOT EXISTS idx_video_watch_sessions_user_id
    ON integration_service.video_watch_sessions(user_id);

CREATE INDEX IF NOT EXISTS idx_video_watch_sessions_video_id
    ON integration_service.video_watch_sessions(video_id);

CREATE INDEX IF NOT EXISTS idx_video_watch_sessions_started_at
    ON integration_service.video_watch_sessions(started_at);