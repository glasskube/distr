INSERT INTO
  Organization (id, name)
VALUES
  (
    '9950bcaf-4fee-4e41-8ad2-eab6083d3c4e',
    'Glasskube'
  );

-- the password for this test user is 12345678
INSERT INTO
  UserAccount (id, email, name, password_hash, password_salt)
VALUES
  (
    '07ad209e-5151-4591-8929-30db4e61f74e',
    'pmig@glasskube.com',
    'Philip Miglinci',
    decode (
      '0284467bc9d5cc09cf6f52d8187133cd2048fc58ad71608272d2b2f63ff10e48',
      'hex'
    ),
    decode ('6299c291d0b08c972e9753684be54c36', 'hex')
  );

INSERT INTO
  Organization_UserAccount (organization_id, user_account_id)
VALUES
  (
    '9950bcaf-4fee-4e41-8ad2-eab6083d3c4e',
    '07ad209e-5151-4591-8929-30db4e61f74e'
  );

INSERT INTO
  Application (id, organization_id, name, type)
VALUES
  (
    '245ebcfe-1c6c-4019-a6be-7037cdd93fc1',
    '9950bcaf-4fee-4e41-8ad2-eab6083d3c4e',
    'ASAN Mars Explorer',
    'docker'
  ),
  (
    '9c3e1886-67e4-47c6-980f-1ab6b5bd622f',
    '9950bcaf-4fee-4e41-8ad2-eab6083d3c4e',
    'Genome Graph Database',
    'docker'
  ),
  (
    '4d5f8130-50fa-49e5-8bca-20590ffd96ca',
    '9950bcaf-4fee-4e41-8ad2-eab6083d3c4e',
    'Wizard Security Graph',
    'docker'
  );

INSERT INTO
  ApplicationVersion (id, name, application_id)
VALUES
  (
    '3bcc5f45-21e7-44a1-83da-8cf880a4bfe5',
    'v4.2.0',
    '245ebcfe-1c6c-4019-a6be-7037cdd93fc1'
  );

INSERT INTO
  ApplicationVersion (name, compose_file_data, application_id)
VALUES
  (
    'v3',
    decode (E 'hello: world\n', 'escape'),
    '14c4daf4-b49b-47c4-a280-e6ca8e33648c'
  ),
  (
    'v2',
    decode (E 'hello: world\n', 'escape'),
    '14c4daf4-b49b-47c4-a280-e6ca8e33648c'
  ),
  (
    'v1',
    decode (E 'hello: world\n', 'escape'),
    '14c4daf4-b49b-47c4-a280-e6ca8e33648c'
  ),
  (
    'v1',
    decode (E 'hello: world\n', 'escape'),
    'a3de7cca-1d6f-4277-9b25-9e502277507f'
  );

INSERT INTO
  DeploymentTarget (
    id,
    organization_id,
    name,
    type,
    geolocation_lat,
    geolocation_lon
  )
VALUES
  (
    'ffb06f6a-ee36-46a4-ba9d-751f19947fff',
    '9950bcaf-4fee-4e41-8ad2-eab6083d3c4e',
    'Space Center Austria',
    'docker',
    48.1956026,
    16.3633028
  ),
  (
    '1f032e69-3249-4bfc-93dc-5e5f4fe7687c',
    '9950bcaf-4fee-4e41-8ad2-eab6083d3c4e',
    'Edge Location',
    'docker',
    NULL,
    NULL
  ),
  (
    'e5cbe194-c7a0-46e1-a987-a7d3ccc1e89f',
    '9950bcaf-4fee-4e41-8ad2-eab6083d3c4e',
    '580 Founders Café',
    'docker',
    37.758781,
    -122.396882
  );

INSERT INTO
  DeploymentTargetStatus (id, deployment_target_id, message)
VALUES
  (
    '46a9a70e-441f-4577-bc52-7664f47f947a',
    'e5cbe194-c7a0-46e1-a987-a7d3ccc1e89f',
    'running'
  );
