import {HttpClient, HttpErrorResponse} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {EMPTY, Observable, retry, shareReplay, switchMap, tap, timer} from 'rxjs';
import {ReactiveList} from './cache';
import {CrudService} from './interfaces';
import {DeploymentTarget, DeploymentTargetAccessResponse} from '@glasskube/cloud-sdk';

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

  private readonly sharedPolling$ = timer(0, 5000).pipe(
    switchMap(() => this.httpClient.get<DeploymentTarget[]>(this.baseUrl)),
    retry({
      delay: (e, c) =>
        e instanceof HttpErrorResponse && (!e.status || e.status >= 500)
          ? timer(Math.min(Math.pow(c, 2), 30) * 1000)
          : EMPTY,
    }),
    shareReplay(1)
  );

  list(): Observable<DeploymentTarget[]> {
    return this.cache.get();
  }

  poll(): Observable<DeploymentTarget[]> {
    return this.sharedPolling$;
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

  requestAccess(deploymentTargetId: string) {
    return this.httpClient.post<DeploymentTargetAccessResponse>(
      `${this.baseUrl}/${deploymentTargetId}/access-request`,
      {}
    );
  }
}
