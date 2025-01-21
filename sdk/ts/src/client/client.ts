import {Application, ApplicationVersion} from '../types/application';
import {DeploymentTarget} from '../types/deployment-target';

export type ClientConfig = {
  apiBase: string;
  apiKey: string;
}

export type ApplicationVersionFiles = {
  composeFile?: string;
  baseValuesFile?: string;
  templateFile?: string;
}

export class Client {
  constructor(private config: ClientConfig) {}

  private async get<T>(path: string): Promise<T> {
    const response = await fetch(`${this.config.apiBase}/${path}`, {
      method: 'GET',
      headers: {
        'Accept': 'application/json',
        'Authorization': `Bearer ${this.config.apiKey}`
      }
    });
    if(response.status < 200 || response.status >= 300) {
      throw new Error(`Failed to GET ${path}: ${response.status} ${response.statusText}`);
    }
    return await response.json() as T;
  }

  private async post<T>(path: string, body: T): Promise<T> {
    const response = await fetch(`${this.config.apiBase}/${path}`, {
      method: 'POST',
      headers: {
        'Accept': 'application/json',
        'Authorization': `Bearer ${this.config.apiKey}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(body)
    });
    if(response.status < 200 || response.status >= 300) {
      throw new Error(`Failed to POST ${path}: ${response.status} ${response.statusText}`);
    }
    return await response.json() as T;
  }

  private async put<T>(path: string, body: T): Promise<T> {
    const response = await fetch(`${this.config.apiBase}/${path}`, {
      method: 'PUT',
      headers: {
        'Accept': 'application/json',
        'Authorization': `Bearer ${this.config.apiKey}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(body)
    });
    if(response.status < 200 || response.status >= 300) {
      throw new Error(`Failed to PUT ${path}: ${response.status} ${response.statusText}`);
    }
    return await response.json() as T;
  }

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

  public async createApplicationVersion(application: Application, version: ApplicationVersion, files?: ApplicationVersionFiles): Promise<Application> {
    const formData = new FormData();
    formData.append('applicationversion', JSON.stringify(version));
    if(files?.composeFile) {
      formData.append('composefile', new Blob([files.composeFile], {type: 'application/yaml'}));
    }
    if(files?.baseValuesFile) {
      formData.append('valuesfile', new Blob([files.baseValuesFile], {type: 'application/yaml'}));
    }
    if(files?.templateFile) {
      formData.append('templatefile', new Blob([files.templateFile], {type: 'application/yaml'}));
    }
    const response = await fetch(`${this.config.apiBase}/applications/${application.id}/versions`, {
      method: 'POST',
      headers: {
        'Accept': 'application/json',
        'Authorization': `Bearer ${this.config.apiKey}`,
      },
      body: formData
    });
    if(response.status < 200 || response.status >= 300) {
      throw new Error(`Failed to create application version: ${response.status} ${response.statusText}: "${await response.text()}"`);
    }
    return await response.json() as Application;
  }

  public async getDeploymentTargets(): Promise<DeploymentTarget[]> {
    return this.get<DeploymentTarget[]>('deployment-targets');
  }

  public async getDeploymentTarget(deploymentTargetId: string): Promise<DeploymentTarget> {
    return this.get<DeploymentTarget>(`deployment-targets/${deploymentTargetId}`);
  }
}
