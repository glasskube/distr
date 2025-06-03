import {HttpClient} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {merge, Observable, shareReplay, Subject, switchMap, tap, timer} from 'rxjs';
import {DefaultReactiveList} from './cache';

interface AgentDeploymentTargetMetrics {
  cpuCoresMillis: number;
  cpuUsage: number;
  memoryBytes: number;
  memoryUsage: number;
}

export interface DeploymentTargetLatestMetrics extends AgentDeploymentTargetMetrics {
  id: string;
}

@Injectable({
  providedIn: 'root',
})
export class DeploymentTargetsMetricsService {
  private readonly deploymentTargetMetricsBaseUrl = '/api/v1/deployment-target-metrics';
  private readonly httpClient = inject(HttpClient);

  private readonly sharedPolling$ = timer(0, 30_000).pipe(
    switchMap(() => this.httpClient.get<DeploymentTargetLatestMetrics[]>(this.deploymentTargetMetricsBaseUrl)),
    shareReplay({
      bufferSize: 1,
      refCount: true,
    })
  );

  poll(): Observable<DeploymentTargetLatestMetrics[]> {
    return this.sharedPolling$;
  }
}
