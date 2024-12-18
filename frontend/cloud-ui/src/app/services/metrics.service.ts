import {inject, Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {Observable} from 'rxjs';
import {UserAccount} from '../types/user-account';

@Injectable({providedIn: 'root'})
export class MetricsService {
  private readonly httpClient = inject(HttpClient);
  private readonly baseUrl = '/api/v1/metrics';

  public getUptimeForDeployment(deploymentId: string): Observable<{total: number, unknown: number}> {
    return this.httpClient.get<{total: number, unknown: number}>(`${this.baseUrl}/uptime?deploymentId=${deploymentId}`);
  }
}
