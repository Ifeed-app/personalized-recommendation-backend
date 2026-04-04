CREATE TABLE IF NOT EXISTS upload_service.uploads (
id UUID PRIMARY KEY,
user_id UUID NOT NULL,
upload_type TEXT NOT NULL,
status TEXT NOT NULL DEFAULT 'pending',
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS upload_service.media_assets (
id UUID PRIMARY KEY,
upload_id UUID NOT NULL,
owner_id UUID NOT NULL,
asset_type TEXT NOT NULL,
storage_key TEXT NOT NULL UNIQUE,
file_name TEXT NOT NULL,
mime_type TEXT NOT NULL,
size_bytes BIGINT NOT NULL,
checksum TEXT,
status TEXT NOT NULL DEFAULT 'uploaded',
created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
CONSTRAINT fk_media_assets_upload
FOREIGN KEY (upload_id)
   REFERENCES upload_service.uploads(id)
   ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_uploads_user_id
    ON upload_service.uploads(user_id);

CREATE INDEX IF NOT EXISTS idx_uploads_status
    ON upload_service.uploads(status);

CREATE INDEX IF NOT EXISTS idx_media_assets_upload_id
    ON upload_service.media_assets(upload_id);

CREATE INDEX IF NOT EXISTS idx_media_assets_owner_id
    ON upload_service.media_assets(owner_id);

CREATE INDEX IF NOT EXISTS idx_media_assets_asset_type
    ON upload_service.media_assets(asset_type);

CREATE INDEX IF NOT EXISTS idx_media_assets_status
    ON upload_service.media_assets(status);