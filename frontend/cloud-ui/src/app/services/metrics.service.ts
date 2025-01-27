import {inject, Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {Observable} from 'rxjs';
import {UptimeMetric} from '@glasskube/cloud-sdk';

@Injectable({providedIn: 'root'})
export class MetricsService {
  private readonly httpClient = inject(HttpClient);
  private readonly baseUrl = '/api/v1/metrics';

  public getUptimeForDeployment(deploymentId: string): Observable<UptimeMetric[]> {
    return this.httpClient.get<UptimeMetric[]>(`${this.baseUrl}/uptime?deploymentId=${deploymentId}`);
  }
}
