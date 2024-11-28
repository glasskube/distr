import {HttpClient, HttpParams} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {Observable, tap} from 'rxjs';
import {DefaultReactiveList} from './cache';
import {Deployment} from '../types/deployment';
import {DeploymentTarget} from '../types/deployment-target';

@Injectable({
  providedIn: 'root',
})
export class DeploymentService {
  private readonly baseUrl = '/api/deployments';
  private readonly httpClient = inject(HttpClient);

  // create(request: Deployment): Observable<Deployment> {
  //   return this.httpClient.post<Deployment>(this.baseUrl, request).pipe(tap((it) => this.cache.save(it)));
  // }

  listForDeploymentTarget(deploymentTargetId: string | undefined): Observable<Deployment[]> {
    if (!deploymentTargetId) {
      return new Observable<Deployment[]>(subscriber => subscriber.next([]));
    }
    const options = {params: new HttpParams().set('deploymentTargetId', deploymentTargetId)};
    return this.httpClient.get<Deployment[]>(this.baseUrl, options);
  }

}
