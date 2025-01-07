CREATE TYPE HELM_CHART_TYPE AS ENUM ('repository', 'oci');

ALTER TABLE ApplicationVersion
  ADD COLUMN chart_type HELM_CHART_TYPE,
  ADD COLUMN chart_name TEXT,
  ADD COLUMN chart_url TEXT,
  ADD COLUMN chart_version TEXT,
  ADD COLUMN values_file_data BYTEA,
  ADD COLUMN template_file_data BYTEA;
