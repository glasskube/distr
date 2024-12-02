INSERT INTO application (id, name, type)
VALUES ('245ebcfe-1c6c-4019-a6be-7037cdd93fc1', 'ASAN Mars Explorer', 'docker'),
       ('9c3e1886-67e4-47c6-980f-1ab6b5bd622f', 'Genome Graph Database', 'docker'),
       ('4d5f8130-50fa-49e5-8bca-20590ffd96ca', 'Wizard Security Graph', 'docker');

INSERT INTO applicationversion (id, name, application_id)
VALUES ('3bcc5f45-21e7-44a1-83da-8cf880a4bfe5', 'v4.2.0', '245ebcfe-1c6c-4019-a6be-7037cdd93fc1');

INSERT INTO applicationversion (name, compose_file_data, application_id)
VALUES ('v3.0.0', decode(E'hello: world\n', 'escape'), '4d5f8130-50fa-49e5-8bca-20590ffd96ca'),
       ('v2.0.0', decode(E'hello: world\n', 'escape'), '4d5f8130-50fa-49e5-8bca-20590ffd96ca'),
       ('v1.0.0', decode(E'hello: world\n', 'escape'), '4d5f8130-50fa-49e5-8bca-20590ffd96ca'),
       ('v1.0.0', decode(E'hello: world\n', 'escape'), '9c3e1886-67e4-47c6-980f-1ab6b5bd622f');

INSERT INTO DeploymentTarget (id, name, type, geolocation_lat, geolocation_lon)
VALUES ('ffb06f6a-ee36-46a4-ba9d-751f19947fff', 'Space Center Austria', 'docker', 48.1956026, 16.3633028),
       ('1f032e69-3249-4bfc-93dc-5e5f4fe7687c', 'Edge Location', 'docker', NULL, NULL),
       ('e5cbe194-c7a0-46e1-a987-a7d3ccc1e89f', '580 Founders Café', 'docker', 37.758781, -122.396882);


INSERT INTO DeploymentTargetStatus (id, deployment_target_id, message)
VALUES ('46a9a70e-441f-4577-bc52-7664f47f947a', 'e5cbe194-c7a0-46e1-a987-a7d3ccc1e89f', 'running'),
       ('46a9a70e-441f-4577-bc52-7664f47f947b', 'ffb06f6a-ee36-46a4-ba9d-751f19947fff', 'running');

INSERT INTO Deployment (deployment_target_id, application_version_id)
VALUES ('ffb06f6a-ee36-46a4-ba9d-751f19947fff', '3bcc5f45-21e7-44a1-83da-8cf880a4bfe5');
