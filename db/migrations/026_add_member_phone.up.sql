ALTER TABLE members
    ADD COLUMN phone VARCHAR(50) NOT NULL DEFAULT '';

CREATE INDEX members_department_phone_idx ON members (department_code, phone);
