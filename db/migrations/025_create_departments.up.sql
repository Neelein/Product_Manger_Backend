CREATE TABLE departments (
    code        INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    CONSTRAINT departments_code_positive CHECK (code > 0),
    CONSTRAINT departments_name_not_blank CHECK (btrim(name) <> '')
);

CREATE UNIQUE INDEX departments_name_trimmed_unique ON departments (btrim(name));

ALTER TABLE members
    ADD COLUMN department_code INTEGER NOT NULL DEFAULT 0,
    ADD CONSTRAINT members_department_code_non_negative CHECK (department_code >= 0);

CREATE INDEX members_employee_search_idx
    ON members (member_type, email, name, id);
