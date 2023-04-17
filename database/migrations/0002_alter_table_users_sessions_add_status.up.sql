CREATE TYPE user_type AS ENUM (
    'authorized',
    'unauthorized',
    'blocked'
    );

ALTER TABLE users ADD COLUMN IF NOT EXISTS type user_type DEFAULT 'authorized';

ALTER TABLE users ADD COLUMN IF NOT EXISTS authentication_times INTEGER;

CREATE TYPE status_type AS ENUM (
    'white_listed',
    'block_listed',
    'warned'
    );

ALTER TABLE users ADD COLUMN IF NOT EXISTS status status_type DEFAULT 'white_listed';



