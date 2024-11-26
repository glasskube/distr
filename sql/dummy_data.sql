INSERT INTO application (name, type)
VALUES ('my-application', 'docker'),
  ('another-application', 'docker'),
  ('application-of-somebody-else', 'docker');

INSERT INTO DeploymentTarget (id, name, type, geolocation_lat, geolocation_lon)
VALUES (
    'ffb06f6a-ee36-46a4-ba9d-751f19947fff',
    'test',
    'docker',
    47.31310,
    14.86250
  ),
  (
    'e5cbe194-c7a0-46e1-a987-a7d3ccc1e89f',
    'test',
    'docker',
    NULL,
    NULL
  );