import {HttpClient, HttpParams} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {ArtifactVersionPull} from '../types/artifact-version-pull';

@Injectable({providedIn: 'root'})
export class ArtifactPullsService {
  private readonly baseUrl = '/api/v1/artifact-pulls';
  private readonly httpClient = inject(HttpClient);

  public get({before, count}: {before?: Date; count?: number} = {}): Observable<ArtifactVersionPull[]> {
    let params = new HttpParams();
    if (before !== undefined) {
      params = params.set('before', before.toISOString());
    }
    if (count !== undefined) {
      params = params.set('count', count);
    }
    return this.httpClient.get<ArtifactVersionPull[]>(this.baseUrl, {params});
  }
}
