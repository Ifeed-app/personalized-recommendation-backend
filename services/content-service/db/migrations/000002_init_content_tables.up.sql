CREATE TABLE IF NOT EXISTS content_service.videos (
id UUID PRIMARY KEY,
creator_id UUID NOT NULL,
title TEXT NOT NULL,
description TEXT,
storage_key TEXT NOT NULL,
thumbnail_key TEXT,
mime_type TEXT NOT NULL,
duration_seconds INTEGER NOT NULL DEFAULT 0,
status TEXT NOT NULL DEFAULT 'draft',
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS content_service.categories (
  id UUID PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS content_service.video_categories (
video_id UUID NOT NULL,
category_id UUID NOT NULL,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
PRIMARY KEY (video_id, category_id),
CONSTRAINT fk_video_categories_video
    FOREIGN KEY (video_id)
        REFERENCES content_service.videos(id)
        ON DELETE CASCADE,
CONSTRAINT fk_video_categories_category
    FOREIGN KEY (category_id)
        REFERENCES content_service.categories(id)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS content_service.tags (
id UUID PRIMARY KEY,
name TEXT NOT NULL UNIQUE,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS content_service.video_tags (
video_id UUID NOT NULL,
tag_id UUID NOT NULL,
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
PRIMARY KEY (video_id, tag_id),
CONSTRAINT fk_video_tags_video
  FOREIGN KEY (video_id)
      REFERENCES content_service.videos(id)
      ON DELETE CASCADE,
CONSTRAINT fk_video_tags_tag
  FOREIGN KEY (tag_id)
      REFERENCES content_service.tags(id)
      ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_videos_creator_id
    ON content_service.videos(creator_id);

CREATE INDEX IF NOT EXISTS idx_videos_status
    ON content_service.videos(status);

CREATE INDEX IF NOT EXISTS idx_categories_name
    ON content_service.categories(name);

CREATE INDEX IF NOT EXISTS idx_tags_name
    ON content_service.tags(name);