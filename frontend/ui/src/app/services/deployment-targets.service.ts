import {HttpClient, HttpErrorResponse} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {
  EMPTY,
  MonoTypeOperatorFunction,
  Observable,
  retry,
  shareReplay,
  Subject,
  switchMap,
  tap,
  timer,
  merge,
} from 'rxjs';
import {ReactiveList} from './cache';
import {CrudService} from './interfaces';
import {Deployment, DeploymentRequest, DeploymentTarget, DeploymentTargetAccessResponse} from '@glasskube/distr-sdk';

class DeploymentTargetsReactiveList extends ReactiveList<DeploymentTarget> {
  protected override identify = (dt: DeploymentTarget) => dt.id;
  protected override sortAttr = (dt: DeploymentTarget) => dt.createdBy?.name ?? dt.createdBy?.email ?? dt.name;
}

@Injectable({
  providedIn: 'root',
})
export class DeploymentTargetsService implements CrudService<DeploymentTarget> {
  private readonly deploymentTargetsBaseUrl = '/api/v1/deployment-targets';
  private readonly deploymentsBaseUrl = '/api/v1/deployments';
  private readonly httpClient = inject(HttpClient);
  private readonly cache = new DeploymentTargetsReactiveList(
    this.httpClient.get<DeploymentTarget[]>(this.deploymentTargetsBaseUrl)
  );

  private readonly pollRefresh$ = new Subject<void>();
  private readonly sharedPolling$ = merge(timer(0, 5000), this.pollRefresh$).pipe(
    switchMap(() => this.httpClient.get<DeploymentTarget[]>(this.deploymentTargetsBaseUrl)),
    tap((dts) => this.cache.save(...dts)),
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
    return this.httpClient.post<DeploymentTarget>(this.deploymentTargetsBaseUrl, request).pipe(
      tap((it) => {
        this.cache.save(it);
        this.pollRefresh$.next();
      })
    );
  }

  update(request: DeploymentTarget): Observable<DeploymentTarget> {
    return this.httpClient.put<DeploymentTarget>(`${this.deploymentTargetsBaseUrl}/${request.id}`, request).pipe(
      tap((it) => {
        this.cache.save(it);
        this.pollRefresh$.next();
      })
    );
  }

  delete(request: DeploymentTarget): Observable<void> {
    return this.httpClient.delete<void>(`${this.deploymentTargetsBaseUrl}/${request.id}`).pipe(
      tap(() => {
        this.cache.remove(request);
        this.pollRefresh$.next();
      })
    );
  }

  requestAccess(deploymentTargetId: string) {
    return this.httpClient.post<DeploymentTargetAccessResponse>(
      `${this.deploymentTargetsBaseUrl}/${deploymentTargetId}/access-request`,
      {}
    );
  }

  deploy(request: DeploymentRequest): Observable<void> {
    return this.httpClient.put<void>(this.deploymentsBaseUrl, request).pipe(tap(() => this.pollRefresh$.next()));
  }

  undeploy(id: string): Observable<void> {
    return this.httpClient.delete<void>(`${this.deploymentsBaseUrl}/${id}`).pipe(tap(() => this.pollRefresh$.next()));
  }
}
