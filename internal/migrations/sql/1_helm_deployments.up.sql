CREATE TYPE HELM_CHART_TYPE AS ENUM ('repository', 'oci');

ALTER TABLE ApplicationVersion ADD COLUMN chart_type HELM_CHART_TYPE;

ALTER TABLE ApplicationVersion ADD COLUMN values_file_data BYTEA;
ALTER TABLE ApplicationVersion ADD COLUMN template_file_data BYTEA;
