import {Client, ClientConfig} from './client';
import {ApplicationVersion} from '../types';

export class CloudService {
  private readonly client: Client;
  constructor(clientConfig: ClientConfig) {
    this.client = new Client(clientConfig);
  }

  public async createDockerApplicationVersion(
    applicationId: string,
    name: string,
    composeFile: string
  ): Promise<ApplicationVersion> {
    return this.client.createApplicationVersion(applicationId, {name}, {composeFile});
  }

  public async createKubernetesApplicationVersion(
    applicationId: string,
    name: string,
    baseValuesFile?: string,
    templateFile?: string
  ): Promise<ApplicationVersion> {
    return this.client.createApplicationVersion(applicationId, {name}, {baseValuesFile, templateFile});
  }
}
