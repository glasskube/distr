import {HttpClient} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {DeploymentLogRecord} from '../types/deployment-log-record';
import {TimeseriesOptions, timeseriesOptionsAsParams} from '../types/timeseries-options';

@Injectable({providedIn: 'root'})
export class DeploymentLogsService {
  private readonly httpClient = inject(HttpClient);

  public getResources(deploymentId: string): Observable<string[]> {
    return this.httpClient.get<string[]>(`/api/v1/deployments/${deploymentId}/logs/resources`);
  }

  public get(deploymentId: string, resource: string, options?: TimeseriesOptions): Observable<DeploymentLogRecord[]> {
    const params = {resource, ...timeseriesOptionsAsParams(options)};
    return this.httpClient.get<DeploymentLogRecord[]>(`/api/v1/deployments/${deploymentId}/logs`, {params});
  }
}
