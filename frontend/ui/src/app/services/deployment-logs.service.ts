import {HttpClient} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {DeploymentLogRecord} from '../types/deployment-log-record';

@Injectable({providedIn: 'root'})
export class DeploymentLogsService {
  private readonly httpClient = inject(HttpClient);

  public getResources(deploymentId: string): Observable<string[]> {
    return this.httpClient.get<string[]>(`/api/v1/deployments/${deploymentId}/logs/resources`);
  }

  public get(deploymentId: string, resource: string, before?: Date): Observable<DeploymentLogRecord[]> {
    const params: Record<string, string> = {resource};
    if (before) {
      params['before'] = before.toISOString();
    }
    return this.httpClient.get<DeploymentLogRecord[]>(`/api/v1/deployments/${deploymentId}/logs`, {params});
  }
}
