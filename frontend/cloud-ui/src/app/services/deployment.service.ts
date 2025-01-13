import {HttpClient, HttpParams} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {Deployment, DeploymentStatus} from '../types/deployment';

@Injectable({
  providedIn: 'root',
})
export class DeploymentService {
  private readonly baseUrl = '/api/v1/deployments';
  private readonly httpClient = inject(HttpClient);

  create(request: Deployment): Observable<Deployment> {
    return this.httpClient.post<Deployment>(this.baseUrl, request);
  }

  getStatuses(depl: Deployment): Observable<DeploymentStatus[]> {
    return this.httpClient.get<DeploymentStatus[]>(`${this.baseUrl}/${depl.id}/status`);
  }
}
