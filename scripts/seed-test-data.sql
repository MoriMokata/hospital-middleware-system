-- Mock/test data for manual dev testing — NOT a migration (migrations/*.sql
-- is embedded and applied automatically on every startup; this file is not,
-- on purpose, since it's fake PII for local testing only).
--
-- Run it against the docker-compose stack, e.g.:
--   docker compose exec -T postgres psql -U postgres -d hospital_middleware \
--     < scripts/seed-test-data.sql
-- or with DBeaver: open this file against the hospital_middleware DB and execute it.
--
-- Re-runnable: every INSERT is guarded so running this twice won't duplicate rows.
--
-- Staff accounts aren't seeded here (passwords need a real bcrypt hash from
-- the app, not something to hand-type into SQL) — create one via the API:
--   curl -X POST http://localhost:8080/staff/create -H "Content-Type: application/json" \
--     -d '{"username":"tester","password":"P@ssw0rd123","hospital":"hospital-a"}'
--   curl -X POST http://localhost:8080/staff/create -H "Content-Type: application/json" \
--     -d '{"username":"tester","password":"P@ssw0rd123","hospital":"hospital-b"}'

-- A second hospital, so cross-hospital isolation can actually be exercised
-- (docker compose's migrations only seed hospital-a).
INSERT INTO hospitals (name, slug, his_adapter_type, his_base_url)
VALUES ('Hospital B', 'hospital-b', 'hospital_b', NULL)
ON CONFLICT (slug) DO NOTHING;

-- Case 1: full record — Thai national, both TH/EN names, national_id, HN, DOB, phone, email, gender F.
INSERT INTO patients (
    hospital_id, patient_hn, national_id,
    first_name, middle_name, last_name,
    first_name_th, middle_name_th, last_name_th,
    first_name_en, middle_name_en, last_name_en,
    date_of_birth, phone_number, email, gender, synced_at
)
SELECT h.id, 'HN-000123', '1234567890123',
    'Somsri', NULL, 'Jaidee',
    'สมศรี', NULL, 'ใจดี',
    'Somsri', NULL, 'Jaidee',
    '1990-05-12', '0812345678', 'somsri@example.com', 'F', now()
FROM hospitals h WHERE h.slug = 'hospital-a'
ON CONFLICT (hospital_id, national_id) WHERE national_id IS NOT NULL DO NOTHING;

-- Case 2: foreign patient — passport_id only (no national_id), gender M.
INSERT INTO patients (
    hospital_id, patient_hn, passport_id,
    first_name, last_name, first_name_en, last_name_en,
    date_of_birth, phone_number, email, gender, synced_at
)
SELECT h.id, 'HN-000124', 'P1234567',
    'John', 'Smith', 'John', 'Smith',
    '1985-11-03', '0898765432', 'john.smith@example.com', 'M', now()
FROM hospitals h WHERE h.slug = 'hospital-a'
ON CONFLICT (hospital_id, passport_id) WHERE passport_id IS NOT NULL DO NOTHING;

-- Case 3: generic name fields only, no TH/EN split — simulates a hospital
-- whose HIS doesn't provide Thai/English-specific fields.
INSERT INTO patients (
    hospital_id, patient_hn, national_id,
    first_name, middle_name, last_name,
    date_of_birth, phone_number, email, gender, synced_at
)
SELECT h.id, 'HN-000125', '2234567890123',
    'Anong', NULL, 'Suksawat',
    '1978-02-20', '0855512345', 'anong.s@example.com', 'F', now()
FROM hospitals h WHERE h.slug = 'hospital-a'
ON CONFLICT (hospital_id, national_id) WHERE national_id IS NOT NULL DO NOTHING;

-- Case 4: minimal record — only national_id + last_name, everything else
-- NULL (tests that nullable columns are handled correctly end to end).
INSERT INTO patients (hospital_id, national_id, last_name, synced_at)
SELECT h.id, '3234567890123', 'Minimal', now()
FROM hospitals h WHERE h.slug = 'hospital-a'
ON CONFLICT (hospital_id, national_id) WHERE national_id IS NOT NULL DO NOTHING;

-- Cases 5-6: similar first/last names, for partial + case-insensitive
-- search matching (e.g. searching "somchai" or "jaidee" should hit both
-- this pair and Case 1 above where relevant).
INSERT INTO patients (
    hospital_id, patient_hn, national_id,
    first_name, last_name, first_name_en, last_name_en,
    date_of_birth, phone_number, gender, synced_at
)
SELECT h.id, 'HN-000126', '4234567890123',
    'Somchai', 'Jaidee', 'Somchai', 'Jaidee',
    '1992-07-15', '0811112222', 'M', now()
FROM hospitals h WHERE h.slug = 'hospital-a'
ON CONFLICT (hospital_id, national_id) WHERE national_id IS NOT NULL DO NOTHING;

INSERT INTO patients (
    hospital_id, patient_hn, national_id,
    first_name, last_name, first_name_en, last_name_en,
    date_of_birth, phone_number, gender, synced_at
)
SELECT h.id, 'HN-000127', '5234567890123',
    'Somchai', 'Jaidum', 'Somchai', 'Jaidum',
    '1995-09-30', '0822223333', 'M', now()
FROM hospitals h WHERE h.slug = 'hospital-a'
ON CONFLICT (hospital_id, national_id) WHERE national_id IS NOT NULL DO NOTHING;

-- Case 7: walk-in with no national_id/passport_id at all — only findable
-- by name/DOB/phone/email, never by an id-based HIS-sync search.
INSERT INTO patients (
    hospital_id, first_name, last_name, date_of_birth, phone_number, gender, synced_at
)
SELECT h.id, 'Walkin', 'NoID', '2000-01-01', '0899998888', 'F', now()
FROM hospitals h WHERE h.slug = 'hospital-a'
AND NOT EXISTS (
    SELECT 1 FROM patients p WHERE p.hospital_id = h.id AND p.first_name = 'Walkin' AND p.last_name = 'NoID'
);

-- Case 8: THE isolation test — a patient in hospital-b with the exact same
-- national_id as Case 1's hospital-a patient. A hospital-a staff member
-- searching by this national_id must never see this row, and vice versa.
INSERT INTO patients (
    hospital_id, patient_hn, national_id,
    first_name, last_name, first_name_en, last_name_en,
    date_of_birth, phone_number, email, gender, synced_at
)
SELECT h.id, 'B-HN-9001', '1234567890123',
    'Different', 'PersonEntirely', 'Different', 'PersonEntirely',
    '1970-03-08', '0866667777', 'different.person@example.com', 'M', now()
FROM hospitals h WHERE h.slug = 'hospital-b'
ON CONFLICT (hospital_id, national_id) WHERE national_id IS NOT NULL DO NOTHING;

-- Case 9: a second hospital-b patient, so hospital-b's search isn't just
-- the isolation-test row.
INSERT INTO patients (
    hospital_id, patient_hn, national_id,
    first_name, last_name, first_name_en, last_name_en,
    date_of_birth, phone_number, email, gender, synced_at
)
SELECT h.id, 'B-HN-9002', '6234567890123',
    'Piya', 'Wongchai', 'Piya', 'Wongchai',
    '1988-12-25', '0877778888', 'piya.w@example.com', 'M', now()
FROM hospitals h WHERE h.slug = 'hospital-b'
ON CONFLICT (hospital_id, national_id) WHERE national_id IS NOT NULL DO NOTHING;
