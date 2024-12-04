INSERT INTO Organization (id, name)
VALUES
  ('9950bcaf-4fee-4e41-8ad2-eab6083d3c4e', 'Glasskube');


INSERT INTO UserAccount (id, email, name, password_hash, password_salt)
VALUES
  -- the password for this test user is 12345678
  (
    '07ad209e-5151-4591-8929-30db4e61f74e',
    'pmig@glasskube.com',
    'Philip Miglinci',
    decode('0284467bc9d5cc09cf6f52d8187133cd2048fc58ad71608272d2b2f63ff10e48', 'hex'),
    decode('6299c291d0b08c972e9753684be54c36', 'hex')
  );


INSERT INTO Organization_UserAccount (organization_id, user_account_id)
VALUES ('9950bcaf-4fee-4e41-8ad2-eab6083d3c4e', '07ad209e-5151-4591-8929-30db4e61f74e');


INSERT INTO application (id, organization_id, name, type)
VALUES ('245ebcfe-1c6c-4019-a6be-7037cdd93fc1', '9950bcaf-4fee-4e41-8ad2-eab6083d3c4e', 'Fastest way to Mars Calculator (Beta)', 'docker'),
       ('9c3e1886-67e4-47c6-980f-1ab6b5bd622f', '9950bcaf-4fee-4e41-8ad2-eab6083d3c4e', 'Fastest way to Mars Calculator (Stable)', 'docker'),
       ('9c3e1886-67e4-47c6-980f-1ab6b5bd622e', '9950bcaf-4fee-4e41-8ad2-eab6083d3c4e', 'Fastest way to Mars Calculator (LTS)', 'kubernetes'),
       ('4d5f8130-50fa-49e5-8bca-20590ffd96ca', '9950bcaf-4fee-4e41-8ad2-eab6083d3c4e', 'Launch Dashboard', 'docker');


INSERT INTO applicationversion (name, compose_file_data, application_id)
VALUES ('v0.1.0', decode(E'hello: world\n', 'escape'), '245ebcfe-1c6c-4019-a6be-7037cdd93fc1'),
       ('v0.2.0', decode(E'hello: world\n', 'escape'), '245ebcfe-1c6c-4019-a6be-7037cdd93fc1'),
       ('v0.3.0', decode(E'hello: world\n', 'escape'), '245ebcfe-1c6c-4019-a6be-7037cdd93fc1'),
       ('v0.3.1', decode(E'hello: world\n', 'escape'), '9c3e1886-67e4-47c6-980f-1ab6b5bd622f');


INSERT INTO applicationversion (id, name, application_id)
VALUES ('3bcc5f45-21e7-44a1-83da-8cf880a4bfe5', 'v4.1.9', '245ebcfe-1c6c-4019-a6be-7037cdd93fc1'),
       ('3bcc5f45-21e7-44a1-83da-8cf880a4bfe4', 'v0.0.1', '4d5f8130-50fa-49e5-8bca-20590ffd96ca'),
       ('3bcc5f45-21e7-44a1-83da-8cf880a4bfe3', 'v0.29.9', '9c3e1886-67e4-47c6-980f-1ab6b5bd622e');


INSERT INTO DeploymentTarget (id, organization_id, name, type, geolocation_lat, geolocation_lon)
VALUES ('ffb06f6a-ee36-46a4-ba9d-751f19947fff', '9950bcaf-4fee-4e41-8ad2-eab6083d3c4e', 'GATE Space', 'docker', 48.191166, 16.3717293),
       ('ffb06f6a-ee36-46a4-ba9d-751f19947ffa', '9950bcaf-4fee-4e41-8ad2-eab6083d3c4e', 'Lumen Orbit', 'docker', 47.6349832, -122.1410062),
       ('1f032e69-3249-4bfc-93dc-5e5f4fe7687c', '9950bcaf-4fee-4e41-8ad2-eab6083d3c4e', 'Alba Orbital', 'kubernetes', 55.8578177, -4.3687363),
       ('e5cbe194-c7a0-46e1-a987-a7d3ccc1e89f', '9950bcaf-4fee-4e41-8ad2-eab6083d3c4e', '580 Founders Café', 'docker', 37.76078, -122.3915258),
       ('e5cbe194-c7a0-46e1-a987-a7d3ccc1e89a', '9950bcaf-4fee-4e41-8ad2-eab6083d3c4e', 'Quindar', 'docker', 39.1929769, -105.2403348);


INSERT INTO DeploymentTargetStatus (id, deployment_target_id, message)
VALUES ('46a9a70e-441f-4577-bc52-7664f47f947a', 'ffb06f6a-ee36-46a4-ba9d-751f19947fff', 'running'),
       ('46a9a70e-441f-4577-bc52-7664f47f947b', 'ffb06f6a-ee36-46a4-ba9d-751f19947ffa', 'running'),
       ('46a9a70e-441f-4577-bc52-7664f47f947c', '1f032e69-3249-4bfc-93dc-5e5f4fe7687c', 'running'),
       ('46a9a70e-441f-4577-bc52-7664f47f947d', 'e5cbe194-c7a0-46e1-a987-a7d3ccc1e89f', 'running');

INSERT INTO Deployment (deployment_target_id, application_version_id)
VALUES ('ffb06f6a-ee36-46a4-ba9d-751f19947fff', '3bcc5f45-21e7-44a1-83da-8cf880a4bfe5'),
       ('ffb06f6a-ee36-46a4-ba9d-751f19947ffa', '3bcc5f45-21e7-44a1-83da-8cf880a4bfe5'),
       ('1f032e69-3249-4bfc-93dc-5e5f4fe7687c', '3bcc5f45-21e7-44a1-83da-8cf880a4bfe3'),
       ('e5cbe194-c7a0-46e1-a987-a7d3ccc1e89f', '3bcc5f45-21e7-44a1-83da-8cf880a4bfe4');
