import {HttpClient} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {Observable, switchMap, timer} from 'rxjs';
import {Deployment, DeploymentRequest, DeploymentRevisionStatus} from '@glasskube/distr-sdk';

@Injectable({
  providedIn: 'root',
})
export class DeploymentService {
  private readonly baseUrl = '/api/v1/deployments';
  private readonly httpClient = inject(HttpClient);

  createOrUpdate(request: DeploymentRequest): Observable<Deployment> {
    return this.httpClient.put<Deployment>(this.baseUrl, request);
  }

  getStatuses(depl: Deployment): Observable<DeploymentRevisionStatus[]> {
    return this.httpClient.get<DeploymentRevisionStatus[]>(`${this.baseUrl}/${depl.id}/status`);
  }

  pollStatuses(depl: Deployment): Observable<DeploymentRevisionStatus[]> {
    return timer(0, 5000).pipe(switchMap(() => this.getStatuses(depl)));
  }
}
