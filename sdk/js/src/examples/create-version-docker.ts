import {DistrService} from '../client/service';
import {clientConfig} from './config';

const gc = new DistrService(clientConfig);
const appId = 'd91ede4e-f909-4f72-a416-f6f5f797682f';

const composeFile = `
services:
  my-postgres:
    image: 'postgres:17.2-alpine3.20'
    ports:
      - '5434:5432'
    environment:
      POSTGRES_USER: \${POSTGRES_USER}
      POSTGRES_PASSWORD: \${POSTGRES_PASSWORD}
      POSTGRES_DB: \${POSTGRES_DB}
    volumes:
      - 'postgres-data:/var/lib/postgresql/data/'

volumes:
  postgres-data:
`

const templateFile = `
POSTGRES_USER=some-user # REPLACE THIS
POSTGRES_PASSWORD=some-password # REPLACE THIS
POSTGRES_DB=some-db # REPLACE THIS`

const newDockerVersion = await gc.createDockerApplicationVersion(
  appId,
  '17.2-alpine3.20+2',
  { composeFile, templateFile }
);
console.log(`* created new version ${newDockerVersion.name} (id: ${newDockerVersion.id}) for docker app ${appId}`);
