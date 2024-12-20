import {HttpClient} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {Observable, tap} from 'rxjs';
import {DeploymentTargetAccessResponse} from '../types/base';
import {DeploymentWithData} from '../types/deployment';
import {DeploymentTarget} from '../types/deployment-target';
import {ReactiveList} from './cache';
import {CrudService} from './interfaces';

class DeploymentTargetsReactiveList extends ReactiveList<DeploymentTarget> {
  protected override identify = (dt: DeploymentTarget) => dt.id;
  protected override sortAttr = (dt: DeploymentTarget) => dt.createdBy?.name ?? dt.createdBy?.email ?? dt.name;
}

@Injectable({
  providedIn: 'root',
})
export class DeploymentTargetsService implements CrudService<DeploymentTarget> {
  private readonly baseUrl = '/api/v1/deployment-targets';
  private readonly httpClient = inject(HttpClient);
  private readonly cache = new DeploymentTargetsReactiveList(this.httpClient.get<DeploymentTarget[]>(this.baseUrl));

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

  delete(request: DeploymentTarget): Observable<void> {
    return this.httpClient.delete<void>(`${this.baseUrl}/${request.id}`).pipe(tap(() => this.cache.remove(request)));
  }

  latestDeploymentFor(deploymentTargetId: string): Observable<DeploymentWithData> {
    return this.httpClient.get<DeploymentWithData>(`${this.baseUrl}/${deploymentTargetId}/latest-deployment`);
  }

  requestAccess(deploymentTargetId: string) {
    return this.httpClient.post<DeploymentTargetAccessResponse>(
      `${this.baseUrl}/${deploymentTargetId}/access-request`,
      {}
    );
  }
}
