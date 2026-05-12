-- Miami-Dade FL: Miami-Dade open data ArcGIS CSV export (bulk building permits).

INSERT INTO county_connectors (
    county_id,
    county_name,
    state,
    url,
    source_type,
    connector_config,
    status,
    updated_at
)
VALUES (
    'miami_dade_fl',
    'Miami-Dade County',
    'FL',
    'https://www.miamidade.gov/permits/search',
    'csv',
    $connector$
{
  "source_type": "csv",
  "api": {
    "endpoint": "https://opendata.arcgis.com/api/v3/datasets/1d6fc60b087c4bcaa22345f429a2ec5a_0/downloads/data?format=csv&spatialRefId=4326",
    "method": "GET"
  },
  "pagination": {"type": "none"},
  "extraction": {
    "fields": [
      {"name": "permit_number", "csv_column": "PermitNumber", "type": "string"},
      {"name": "application_number", "csv_column": "ApplicationNumber", "type": "string"},
      {"name": "process_number", "csv_column": "ProcessNumber", "type": "string"},
      {"name": "issued_date", "csv_column": "IssuedDate", "type": "string"},
      {"name": "first_submission_date", "csv_column": "FirstSubmissionDate", "type": "string"},
      {"name": "plan_accepted_date", "csv_column": "PlanAcceptedDate", "type": "string"},
      {"name": "building_permit_status", "csv_column": "BuildingPermitStatusDescription", "type": "string"},
      {"name": "scope_of_work", "csv_column": "ScopeofWork", "type": "string"},
      {"name": "folio_number", "csv_column": "FolioNumber", "type": "string"},
      {"name": "latitude", "csv_column": "Latitude", "type": "string"},
      {"name": "longitude", "csv_column": "Longitude", "type": "string"},
      {"name": "property_type", "csv_column": "PropertyType", "type": "string"},
      {"name": "total_cost", "csv_column": "TotalCost", "type": "string"},
      {"name": "remodeling_cost", "csv_column": "RemodelingCost", "type": "string"},
      {"name": "new_addition_cost", "csv_column": "NewAdditionCost", "type": "string"},
      {"name": "delivery_address", "csv_column": "DeliveryAddress", "type": "string"},
      {"name": "company_name", "csv_column": "CompanyName", "type": "string"},
      {"name": "work_items", "csv_column": "WorkItems", "type": "string"}
    ]
  },
  "dedup": {"unique_field": "permit_number"},
  "rate_limit": {}
}
$connector$::jsonb,
    'active',
    NOW()
)
ON CONFLICT (county_id) DO UPDATE SET
    connector_config = EXCLUDED.connector_config,
    source_type = EXCLUDED.source_type,
    url = EXCLUDED.url,
    county_name = EXCLUDED.county_name,
    state = EXCLUDED.state,
    status = EXCLUDED.status,
    updated_at = NOW();
