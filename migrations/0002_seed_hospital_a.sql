-- Task 03: seed Hospital A so staff/create and the Hospital A adapter have a row to reference
INSERT INTO hospitals (name, slug, his_adapter_type, his_base_url)
VALUES ('Hospital A', 'hospital-a', 'hospital_a', 'https://hospital-a.api.co.th')
ON CONFLICT (slug) DO NOTHING;
