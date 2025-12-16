import {HttpClient} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {Secret} from '../types/secret';

const baseUrl = '/api/v1/secrets';

@Injectable({providedIn: 'root'})
export class SecretsService {
  private readonly httpClient = inject(HttpClient);

  public list(): Observable<Secret[]> {
    return this.httpClient.get<Secret[]>(baseUrl);
  }

  public put(key: string, value: string): Observable<Secret> {
    return this.httpClient.put<Secret>(`${baseUrl}/${key}`, {value});
  }

  public delete(key: string): Observable<void> {
    return this.httpClient.delete<void>(`${baseUrl}/${key}`);
  }
}
