INSERT INTO application (id, name, type)
VALUES (
    '14c4daf4-b49b-47c4-a280-e6ca8e33648c',
    'my-application',
    'docker'
  ),
  (
    'a3de7cca-1d6f-4277-9b25-9e502277507f',
    'another-application',
    'docker'
  ),
  (
    '9f842b7b-d9ee-49fc-aa4f-5006d840d53a',
    'application-of-somebody-else',
    'docker'
  );

INSERT INTO applicationversion (name, compose_file_data, application_id)
VALUES ('v3', decode(E'hello: world\n', 'escape'), '14c4daf4-b49b-47c4-a280-e6ca8e33648c'),
       ('v2', decode(E'hello: world\n', 'escape'), '14c4daf4-b49b-47c4-a280-e6ca8e33648c'),
       ('v1', decode(E'hello: world\n', 'escape'), '14c4daf4-b49b-47c4-a280-e6ca8e33648c'),
       ('v1', decode(E'hello: world\n', 'escape'), 'a3de7cca-1d6f-4277-9b25-9e502277507f');

INSERT INTO DeploymentTarget (id, name, type, geolocation_lat, geolocation_lon)
VALUES (
    'ffb06f6a-ee36-46a4-ba9d-751f19947fff',
    'Vienna Office',
    'docker',
    48.1956026,
    16.3633028
  ),
  (
    '1f032e69-3249-4bfc-93dc-5e5f4fe7687c',
    'Secret customer',
    'docker',
    NULL,
    NULL
  ),
  (
    'e5cbe194-c7a0-46e1-a987-a7d3ccc1e89f',
    'San Francisco Office',
    'docker',
    37.758781,
    -122.396882
  );

INSERT INTO DeploymentTargetStatus (id, deployment_target_id, message)
VALUES(
    '46a9a70e-441f-4577-bc52-7664f47f947a',
    'e5cbe194-c7a0-46e1-a987-a7d3ccc1e89f',
    'running'
  )
