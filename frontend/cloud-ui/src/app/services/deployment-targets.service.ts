import {HttpClient} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {CrudService} from './interfaces';
import {DeploymentTarget} from '../types/deployment-target';
import {Observable, tap} from 'rxjs';
import {DefaultReactiveList} from './cache';
import {DeploymentWithData} from '../types/deployment';

@Injectable({
  providedIn: 'root',
})
export class DeploymentTargetsService implements CrudService<DeploymentTarget> {
  private readonly baseUrl = '/api/deployment-targets';
  private readonly httpClient = inject(HttpClient);
  private readonly cache = new DefaultReactiveList(this.httpClient.get<DeploymentTarget[]>(this.baseUrl));

  list(): Observable<DeploymentTarget[]> {
    return this.cache.get();
  }

  create(request: DeploymentTarget): Observable<DeploymentTarget> {
    return this.httpClient.post<DeploymentTarget>(this.baseUrl, request).pipe(tap((it) => this.cache.save(it)));
  }

  update(request: DeploymentTarget): Observable<DeploymentTarget> {
    return this.httpClient
      .put<DeploymentTarget>(`${this.baseUrl}/${request.id}`, request)
      .pipe(tap((it) => this.cache.save(it)));
  }

  latestDeploymentFor(deploymentTargetId: string): Observable<DeploymentWithData> {
    return this.httpClient.get<DeploymentWithData>(`${this.baseUrl}/${deploymentTargetId}/latest-deployment`);
  }
}
