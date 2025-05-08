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
import {
  Deployment,
  DeploymentRequest,
  DeploymentTarget,
  DeploymentTargetAccessResponse,
  DeploymentTargetBase,
} from '@glasskube/distr-sdk';

class DeploymentTargetsReactiveList extends ReactiveList<DeploymentTargetMetrics> {
  protected override identify = (dt: DeploymentTargetMetrics) => dt.id;
  protected override sortAttr = (dt: DeploymentTargetMetrics) => dt.createdBy?.name ?? dt.createdBy?.email ?? dt.name;
}

interface AgentDeploymentTargetMetrics {
  cpuCoresM: number;
  cpuUsage: number;
  memoryBytes: number;
  memoryUsage: number;
}

export interface DeploymentTargetMetrics extends DeploymentTargetBase, AgentDeploymentTargetMetrics {}

@Injectable({
  providedIn: 'root',
})
export class DeploymentTargetsMetricsService {
  private readonly deploymentTargetMetricsBaseUrl = '/api/v1/deployment-target-metrics';
  private readonly httpClient = inject(HttpClient);
  private readonly cache = new DeploymentTargetsReactiveList(
    this.httpClient.get<DeploymentTargetMetrics[]>(this.deploymentTargetMetricsBaseUrl)
  );

  private readonly pollRefresh$ = new Subject<void>();
  private readonly sharedPolling$ = merge(timer(0, 30_000), this.pollRefresh$).pipe(
    switchMap(() => this.httpClient.get<DeploymentTargetMetrics[]>(this.deploymentTargetMetricsBaseUrl)),
    tap((dts) => this.cache.reset(dts)),
    shareReplay({
      bufferSize: 1,
      refCount: true,
    })
  );

  list(): Observable<DeploymentTargetMetrics[]> {
    return this.cache.get();
  }

  poll(): Observable<DeploymentTargetMetrics[]> {
    return this.sharedPolling$;
  }
}
