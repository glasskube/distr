import {HttpClient} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable, of, tap} from 'rxjs';
import {Organization} from '../types/organization';

@Injectable({
  providedIn: 'root',
})
export class OrganizationService {
  private readonly baseUrl = '/api/v1/organization';
  private cache?: Organization;

  constructor(private readonly httpClient: HttpClient) {}

  get(): Observable<Organization> {
    if (this.cache) {
      return of(this.cache);
    }
    return this.httpClient.get<Organization>(this.baseUrl).pipe(tap((org) => (this.cache = org)));
  }
}
