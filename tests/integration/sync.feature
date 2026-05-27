Feature: Product Synchronization
  In order to have up-to-date product data in Elasticsearch
  As a product sync pipeline
  I need to build an Elasticsearch document with translations and handle fallbacks/caching

  Background:
    Given a clean PostgreSQL database with the product schema
    And the following products exist:
      | id                                   | sku            | part_number | brand |
      | 00000000-0000-0000-0000-000000000001 | BP-OIL-5W30-1L | 5W30-1L     | bosch |
    And the following attributes exist:
      | id                                   | code      | metric_unit |
      | 10000000-0000-0000-0000-000000000001 | oil_grade |             |
    And the following specifications exist:
      | id                                   | product_id                           | attribute_id                         | value |
      | 20000000-0000-0000-0000-000000000001 | 00000000-0000-0000-0000-000000000001 | 10000000-0000-0000-0000-000000000001 | 5w30  |

  Scenario: Successful product document synchronization with translations
    Given the following translations exist:
      | entity_type            | entity_id                            | locale | field_name  | field_value                |
      | product                | 00000000-0000-0000-0000-000000000001 | en     | productname | 5W-30 Engine Oil 1L        |
      | product                | 00000000-0000-0000-0000-000000000001 | th     | productname | น้ำมันเครื่อง 5W-30 1 ลิตร |
      | product_specification  | 20000000-0000-0000-0000-000000000001 | en     | value_label | 5W-30                      |
    When I build the document for product "00000000-0000-0000-0000-000000000001" with locales "en,th"
    Then the document SKU should be "BP-OIL-5W30-1L"
    And the document should contain the English product name "5W-30 Engine Oil 1L"
    And the document should contain the Thai product name "น้ำมันเครื่อง 5W-30 1 ลิตร"
    And the document's oil_grade English label should be "5W-30"

  Scenario: Fallback to default locale for missing translations
    Given the following translations exist:
      | entity_type | entity_id                            | locale | field_name  | field_value           |
      | product     | 00000000-0000-0000-0000-000000000001 | en     | productname | 5W-30 Engine Oil 1L   |
    When I build the document for product "00000000-0000-0000-0000-000000000001" with locales "en,th"
    Then the document should contain the Thai product name "5W-30 Engine Oil 1L"

  Scenario: Synchronization picks up new data after cache invalidation
    Given the following translations exist:
      | entity_type | entity_id                            | locale | field_name  | field_value |
      | product     | 00000000-0000-0000-0000-000000000001 | en     | productname | Old Name    |
    When I build the document for product "00000000-0000-0000-0000-000000000001" with locales "en"
    And the document should contain the English product name "Old Name"
    When the translation for product "00000000-0000-0000-0000-000000000001" (locale "en", field "productname") is updated to "New Name" in the database
    And I invalidate the cache for entity "00000000-0000-0000-0000-000000000001"
    And I build the document for product "00000000-0000-0000-0000-000000000001" with locales "en"
    Then the document should contain the English product name "New Name"

  Scenario: Handle non-existent product gracefully
    When I build the document for product "00000000-0000-0000-0000-999999999999" with locales "en"
    Then the document build should fail

  Scenario: Second sync for the same product uses the translation cache
    Given a product exists with SKU "BP-CACHE-HIT" and the following translations:
      | locale | field_name  | field_value       |
      | en     | productname | Cache Hit Product |
      | th     | productname | สินค้าแคช         |
    When I sync the product with locales "en,th"
    Then the document should contain the English product name "Cache Hit Product"
    When all translations for the product are deleted from the database
    And I sync the product again with locales "en,th"
    Then the document should contain the English product name "Cache Hit Product"

  Scenario: Partial cache hit for a new locale forces a fresh database fetch
    Given a product exists with SKU "BP-PARTIAL" and the following translations:
      | locale | field_name  | field_value |
      | en     | productname | Partial EN  |
    When I sync the product with locales "en"
    Then the document should contain the English product name "Partial EN"
    Given a Thai translation is added for the product:
      | locale | field_name  | field_value |
      | th     | productname | บางส่วน TH  |
    When I sync the product with locales "en,th"
    Then the document should contain the English product name "Partial EN"
    And the document should contain the Thai product name "บางส่วน TH"

  Scenario: Product with no translations uses SKU and raw spec values as fallback
    Given a product exists with SKU "BP-NOTRANS" and no translations
    When I sync the product with locales "en,th"
    Then the document SKU should be "BP-NOTRANS"
    And the document should contain the English product name "BP-NOTRANS"
    And the document should contain the Thai product name "BP-NOTRANS"
