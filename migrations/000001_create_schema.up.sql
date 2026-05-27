CREATE TABLE product (id UUID PRIMARY KEY, sku VARCHAR, part_number VARCHAR, brand VARCHAR, category_id UUID);
CREATE TABLE attribute (id UUID PRIMARY KEY, code VARCHAR, metric_unit VARCHAR);
CREATE TABLE product_specification (id UUID PRIMARY KEY, product_id UUID REFERENCES product(id), attribute_id UUID REFERENCES attribute(id), value VARCHAR);
CREATE TABLE translation (id UUID PRIMARY KEY DEFAULT gen_random_uuid(), entity_type VARCHAR, entity_id VARCHAR, locale VARCHAR, field_name VARCHAR, field_value TEXT, updated_at TIMESTAMPTZ DEFAULT NOW());
