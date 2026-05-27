-- fixtures/data.sql

-- Truncate tables to ensure a clean state
TRUNCATE TABLE product, attribute, product_specification, translation RESTART IDENTITY CASCADE;

-- Insert sample product
INSERT INTO product (id, sku, part_number, brand, category_id)
VALUES ('00000000-0000-0000-0000-000000000001', 'SKU-001', 'PN-001', 'Brand-A', NULL);

-- Insert sample attribute
INSERT INTO attribute (id, code, metric_unit)
VALUES ('00000000-0000-0000-0000-000000000002', 'oil_grade', 'viscosity');

-- Insert sample specification
INSERT INTO product_specification (id, product_id, attribute_id, value)
VALUES ('00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000002', '5W-30');

-- Insert sample translation
INSERT INTO translation (entity_type, entity_id, locale, field_name, field_value)
VALUES ('product', '00000000-0000-0000-0000-000000000001', 'en', 'product_name', 'Super Engine Oil'),
       ('product', '00000000-0000-0000-0000-000000000001', 'th', 'product_name', 'น้ำมันเครื่อง ซูเปอร์');
