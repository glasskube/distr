import {HttpClient} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {merge, Observable, shareReplay, Subject, tap} from 'rxjs';
import {Organization} from '../types/organization';

@Injectable({
  providedIn: 'root',
})
export class OrganizationService {
  private readonly httpClient = inject(HttpClient);
  private readonly baseUrl = '/api/v1/organization';
  private readonly organizationUpdate = new Subject<Organization>();
  private readonly organization$ = merge(
    this.organizationUpdate.asObservable(),
    this.httpClient.get<Organization>(this.baseUrl)
  ).pipe(shareReplay(1));

  get(): Observable<Organization> {
    return this.organization$.pipe(shareReplay(1));
  }

  update(organization: Organization): Observable<Organization> {
    return this.httpClient
      .put<Organization>(this.baseUrl, organization)
      .pipe(tap((it) => this.organizationUpdate.next(it)));
  }
}
