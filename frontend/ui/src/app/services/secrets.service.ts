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

  public create(key: string, value: string, customerOrganizationId?: string): Observable<Secret> {
    return this.httpClient.post<Secret>(baseUrl, {key, value, customerOrganizationId});
  }

  public update(id: string, value: string): Observable<Secret> {
    return this.httpClient.put<Secret>(`${baseUrl}/${id}`, {value});
  }

  public delete(id: string): Observable<void> {
    return this.httpClient.delete<void>(`${baseUrl}/${id}`);
  }
}
