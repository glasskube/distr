INSERT INTO application (id, name, type)
VALUES ('245ebcfe-1c6c-4019-a6be-7037cdd93fc1', 'ASAN Mars Explorer', 'docker'),
       ('9c3e1886-67e4-47c6-980f-1ab6b5bd622f', 'Genome Graph Database', 'docker'),
       ('4d5f8130-50fa-49e5-8bca-20590ffd96ca', 'Wizard Security Graph', 'docker');

INSERT INTO applicationversion (id, name, application_id)
VALUES ('3bcc5f45-21e7-44a1-83da-8cf880a4bfe5', 'v4.2.0', '245ebcfe-1c6c-4019-a6be-7037cdd93fc1');

INSERT INTO deploymenttarget (id, name, type, geolocation_lat, geolocation_lon)
VALUES ('ffb06f6a-ee36-46a4-ba9d-751f19947fff', 'Kurz Space Center Austria', 'docker', 47.31310, 14.86250),
       ('e5cbe194-c7a0-46e1-a987-a7d3ccc1e89f', '580 Founders Café', 'docker', NULL, NULL);

INSERT INTO deployment (deployment_target_id, application_version_id)
VALUES ('ffb06f6a-ee36-46a4-ba9d-751f19947fff', '3bcc5f45-21e7-44a1-83da-8cf880a4bfe5')
