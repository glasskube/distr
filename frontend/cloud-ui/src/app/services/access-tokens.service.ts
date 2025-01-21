import {HttpClient} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {AccessToken, AccessTokenWithKey, CreateAccessTokenRequest} from '../types/access-token';

const baseUrl = '/api/v1/settings/tokens';

@Injectable({providedIn: 'root'})
export class AccessTokensService {
  private readonly httpClient = inject(HttpClient);

  public list(): Observable<AccessToken[]> {
    return this.httpClient.get<AccessToken[]>(baseUrl);
  }

  public create(request: CreateAccessTokenRequest): Observable<AccessTokenWithKey> {
    return this.httpClient.post<AccessTokenWithKey>(baseUrl, request);
  }

  public delete(id: string): Observable<void> {
    return this.httpClient.delete<void>(`${baseUrl}/${id}`);
  }
}
