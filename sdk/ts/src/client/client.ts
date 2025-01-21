import {Application, ApplicationVersion, DeploymentRequest, DeploymentTarget} from '../types';

export type ClientConfig = {
  apiBase: string;
  apiKey: string;
};

export type ApplicationVersionFiles = {
  composeFile?: string;
  baseValuesFile?: string;
  templateFile?: string;
};

export class Client {
  constructor(private readonly config: ClientConfig) {}

  public async getApplications(): Promise<Application[]> {
    return this.get<Application[]>('applications');
  }

  public async getApplication(applicationId: string): Promise<Application> {
    return this.get<Application>(`applications/${applicationId}`);
  }

  public async createApplication(application: Application): Promise<Application> {
    return this.post<Application>('applications', application);
  }

  public async updateApplication(application: Application): Promise<Application> {
    return this.put<Application>(`applications/${application.id}`, application);
  }

  public async createApplicationVersion(
    applicationId: string,
    version: ApplicationVersion,
    files?: ApplicationVersionFiles
  ): Promise<ApplicationVersion> {
    const formData = new FormData();
    formData.append('applicationversion', JSON.stringify(version));
    if (files?.composeFile) {
      formData.append('composefile', new Blob([files.composeFile], {type: 'application/yaml'}));
    }
    if (files?.baseValuesFile) {
      formData.append('valuesfile', new Blob([files.baseValuesFile], {type: 'application/yaml'}));
    }
    if (files?.templateFile) {
      formData.append('templatefile', new Blob([files.templateFile], {type: 'application/yaml'}));
    }
    const path = `applications/${applicationId}/versions`;
    const response = await fetch(`${this.config.apiBase}/${path}`, {
      method: 'POST',
      headers: {
        Accept: 'application/json',
        Authorization: `Bearer ${this.config.apiKey}`,
      },
      body: formData,
    });
    return this.handleResponse<ApplicationVersion>(response, 'POST', path);
  }

  public async getDeploymentTargets(): Promise<DeploymentTarget[]> {
    return this.get<DeploymentTarget[]>('deployment-targets');
  }

  public async getDeploymentTarget(deploymentTargetId: string): Promise<DeploymentTarget> {
    return this.get<DeploymentTarget>(`deployment-targets/${deploymentTargetId}`);
  }

  public async createDeploymentTarget(deploymentTarget: DeploymentTarget): Promise<DeploymentTarget> {
    return this.post<DeploymentTarget>('deployment-targets', deploymentTarget);
  }

  public async createOrUpdateDeployment(deploymentRequest: DeploymentRequest): Promise<DeploymentRequest> {
    return this.put<DeploymentRequest>('deployments', deploymentRequest);
  }

  private async get<T>(path: string): Promise<T> {
    const response = await fetch(`${this.config.apiBase}/${path}`, {
      method: 'GET',
      headers: {
        Accept: 'application/json',
        Authorization: `Bearer ${this.config.apiKey}`,
      },
    });
    return await this.handleResponse<T>(response, 'GET', path);
  }

  private async post<T>(path: string, body: T): Promise<T> {
    const response = await fetch(`${this.config.apiBase}/${path}`, {
      method: 'POST',
      headers: {
        Accept: 'application/json',
        Authorization: `Bearer ${this.config.apiKey}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(body),
    });
    return await this.handleResponse<T>(response, 'POST', path);
  }

  private async put<T>(path: string, body: T): Promise<T> {
    const response = await fetch(`${this.config.apiBase}/${path}`, {
      method: 'PUT',
      headers: {
        Accept: 'application/json',
        Authorization: `Bearer ${this.config.apiKey}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(body),
    });
    return await this.handleResponse<T>(response, 'PUT', path);
  }

  private async handleResponse<T>(response: Response, method: string, path: string) {
    if (response.status < 200 || response.status >= 300) {
      throw new Error(`${method} ${path} failed: ${response.status} ${response.statusText} "${await response.text()}"`);
    }
    return (await response.json()) as T;
  }
}
