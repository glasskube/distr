import {HttpClient} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {DeploymentRevisionStatus} from '@glasskube/distr-sdk';
import {Observable, switchMap, timer} from 'rxjs';
import {TimeseriesOptions, timeseriesOptionsAsParams} from '../types/timeseries-options';

@Injectable({
  providedIn: 'root',
})
export class DeploymentStatusService {
  private readonly baseUrl = '/api/v1/deployments';
  private readonly httpClient = inject(HttpClient);

  getStatuses(deploymentId: string, options?: TimeseriesOptions): Observable<DeploymentRevisionStatus[]> {
    const params = timeseriesOptionsAsParams(options);
    return this.httpClient.get<DeploymentRevisionStatus[]>(`${this.baseUrl}/${deploymentId}/status`, {params});
  }

  pollStatuses(deploymentId: string): Observable<DeploymentRevisionStatus[]> {
    return timer(0, 5000).pipe(switchMap(() => this.getStatuses(deploymentId)));
  }
}
