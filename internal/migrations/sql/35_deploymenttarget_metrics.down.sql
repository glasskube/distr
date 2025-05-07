ALTER TABLE DeploymentTarget DROP COLUMN metrics_enabled;

DROP TABLE IF EXISTS DeploymentTargetMetrics;
