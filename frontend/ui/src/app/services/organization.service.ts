import {HttpClient} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {merge, Observable, shareReplay, Subject, tap} from 'rxjs';
import {Organization, OrganizationWithUserRole} from '../types/organization';

@Injectable({
  providedIn: 'root',
})
export class OrganizationService {
  private readonly httpClient = inject(HttpClient);
  private readonly baseUrl = '/api/v1/organization';
  private readonly listUrl = '/api/v1/organizations';

  private readonly organizationUpdate = new Subject<Organization>();
  private readonly organization$ = merge(
    this.organizationUpdate.asObservable(),
    this.httpClient.get<Organization>(this.baseUrl)
  ).pipe(shareReplay(1));
  private readonly organizations$ = this.httpClient.get<OrganizationWithUserRole[]>(this.listUrl).pipe(shareReplay(1));

  get(): Observable<Organization> {
    return this.organization$.pipe(shareReplay(1));
  }

  list(): Observable<OrganizationWithUserRole[]> {
    return this.organizations$;
  }

  update(organization: Organization): Observable<Organization> {
    return this.httpClient
      .put<Organization>(this.baseUrl, organization)
      .pipe(tap((it) => this.organizationUpdate.next(it)));
  }
}
