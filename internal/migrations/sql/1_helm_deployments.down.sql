ALTER TABLE ApplicationVersion DROP COLUMN values_file_data;
ALTER TABLE ApplicationVersion DROP COLUMN template_file_data;

DROP TYPE IF EXISTS HELM_CHART_TYPE CASCADE;
