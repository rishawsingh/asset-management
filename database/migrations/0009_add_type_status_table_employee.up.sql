CREATE TYPE employee_status AS ENUM (
    'active',
    'deleted',
    'not_an_employee'
    );
CREATE TYPE employee_type AS ENUM (
    'employee',
    'intern',
    'freelancer'
    );

ALTER TABLE employee ADD COLUMN IF NOT EXISTS status employee_status DEFAULT 'active';
ALTER TABLE employee ADD COLUMN IF NOT EXISTS type employee_type;