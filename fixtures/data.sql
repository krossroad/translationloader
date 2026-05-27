-- fixtures/data.sql

-- Truncate tables to ensure a clean state
TRUNCATE TABLE product, attribute, product_specification, translation RESTART IDENTITY CASCADE;

-- Insert sample products
INSERT INTO product (id, sku, part_number, brand, category_id)
VALUES
('00000000-0000-0000-0000-000000000001', 'SKU-001', 'PN-001', 'Bosch', NULL),
('00000000-0000-0000-0000-000000000002', 'SKU-002', 'PN-002', 'Castrol', NULL),
('00000000-0000-0000-0000-000000000003', 'SKU-003', 'PN-003', 'Mobil1', NULL);

-- Insert sample attributes
INSERT INTO attribute (id, code, metric_unit)
VALUES
('00000000-0000-0000-0000-000000000010', 'oil_grade', 'viscosity'),
('00000000-0000-0000-0000-000000000011', 'volume', 'liters');

-- Insert sample specifications
INSERT INTO product_specification (id, product_id, attribute_id, value)
VALUES
('00000000-0000-0000-0000-000000000020', '00000000-0000-0000-0000-000000000001', '00000000-0000-0000-0000-000000000010', '5W-30'),
('00000000-0000-0000-0000-000000000021', '00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000010', '10W-40'),
('00000000-0000-0000-0000-000000000022', '00000000-0000-0000-0000-000000000003', '00000000-0000-0000-0000-000000000011', '5');

-- Insert sample translations
INSERT INTO translation (entity_type, entity_id, locale, field_name, field_value)
VALUES
-- Bosch Product
('product', '00000000-0000-0000-0000-000000000001', 'en', 'productname', 'Bosch Synthetic Oil 5W-30'),
('product', '00000000-0000-0000-0000-000000000001', 'th', 'productname', 'น้ำมันเครื่อง บ๊อช ซินเธติก'),
-- Castrol Product
('product', '00000000-0000-0000-0000-000000000002', 'en', 'productname', 'Castrol Magnatec 10W-40'),
('product', '00000000-0000-0000-0000-000000000002', 'th', 'productname', 'น้ำมันเครื่อง คาสตรอล แมกนาเทค'),
-- Mobil1 Product
('product', '00000000-0000-0000-0000-000000000003', 'en', 'productname', 'Mobil1 Full Synthetic 5L'),
('product', '00000000-0000-0000-0000-000000000003', 'th', 'productname', 'น้ำมันเครื่อง โมบิลวัน');
