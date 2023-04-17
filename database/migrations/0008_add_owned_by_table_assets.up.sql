CREATE TYPE asset_owned_status AS ENUM (
    'remote_state',
    'client'
    );
CREATE TYPE asset_status AS ENUM (
    'available',
    'assigned',
    'deleted'
    );

ALTER TABLE assets ADD COLUMN IF NOT EXISTS owned_by asset_owned_status;

ALTER TABLE assets ADD COLUMN IF NOT EXISTS client_name TEXT;

ALTER TABLE assets ADD COLUMN IF NOT EXISTS status asset_status DEFAULT 'available';

ALTER TABLE assets ADD COLUMN IF NOT EXISTS deleted_by UUID REFERENCES users(id);

ALTER TABLE assets ADD COLUMN IF NOT EXISTS archive_reason TEXT;