import {HttpClient} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {Observable, of, shareReplay, tap} from 'rxjs';
import {Organization} from '../types/organization';

@Injectable({
  providedIn: 'root',
})
export class OrganizationService {
  private readonly httpClient = inject(HttpClient);
  private readonly baseUrl = '/api/v1/organization';
  private readonly organization$ = this.httpClient.get<Organization>(this.baseUrl).pipe(shareReplay(1));

  get(): Observable<Organization> {
    return this.organization$;
  }
}
