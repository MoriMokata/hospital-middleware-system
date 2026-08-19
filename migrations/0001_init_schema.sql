-- Task 03: initial schema — hospitals, staff, patients (see docs/er-diagram.md)
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE hospitals (
    id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name             TEXT NOT NULL,
    slug             TEXT NOT NULL UNIQUE,
    his_adapter_type TEXT NOT NULL,
    his_base_url     TEXT,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE staff (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hospital_id   UUID NOT NULL REFERENCES hospitals(id),
    username      TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (hospital_id, username)
);

CREATE TABLE patients (
    id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    hospital_id    UUID NOT NULL REFERENCES hospitals(id),
    patient_hn     TEXT,
    national_id    TEXT,
    passport_id    TEXT,
    first_name     TEXT,
    middle_name    TEXT,
    last_name      TEXT,
    first_name_th  TEXT,
    middle_name_th TEXT,
    last_name_th   TEXT,
    first_name_en  TEXT,
    middle_name_en TEXT,
    last_name_en   TEXT,
    date_of_birth  DATE,
    phone_number   TEXT,
    email          TEXT,
    gender         TEXT,
    raw_source     JSONB,
    synced_at      TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE UNIQUE INDEX uq_patients_hospital_national_id
    ON patients (hospital_id, national_id) WHERE national_id IS NOT NULL;
CREATE UNIQUE INDEX uq_patients_hospital_passport_id
    ON patients (hospital_id, passport_id) WHERE passport_id IS NOT NULL;
CREATE INDEX idx_patients_hospital_phone ON patients (hospital_id, phone_number);
CREATE INDEX idx_patients_hospital_email ON patients (hospital_id, email);
CREATE INDEX idx_patients_hospital_name ON patients (hospital_id, last_name, first_name);
