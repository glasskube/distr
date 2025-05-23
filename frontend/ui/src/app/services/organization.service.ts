import {HttpClient} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {combineLatestWith, map, merge, Observable, shareReplay, Subject, tap} from 'rxjs';
import {Organization, OrganizationWithUserRole} from '../types/organization';
import {ContextService} from './context.service';

@Injectable({
  providedIn: 'root',
})
export class OrganizationService {
  private readonly httpClient = inject(HttpClient);
  private readonly contextService = inject(ContextService);
  private readonly baseUrl = '/api/v1/organization';

  private readonly organizationUpdate = new Subject<OrganizationWithUserRole>();
  private readonly organization$ = merge(
    this.organizationUpdate.asObservable(),
    this.contextService.getOrganization()
  ).pipe(shareReplay(1));

  get(): Observable<OrganizationWithUserRole> {
    return this.organization$.pipe(shareReplay(1));
  }

  getAll(): Observable<OrganizationWithUserRole[]> {
    // TODO take updates into account like with organization$
    return this.contextService.getAvailableOrganizations();
  }

  update(organization: Organization): Observable<Organization> {
    return this.httpClient.put<Organization>(this.baseUrl, organization).pipe(
      combineLatestWith(this.getAll()),
      map(([it, allOrgs]) => {
        const foundOrg = allOrgs.find((o) => o.id === it.id);
        return {
          ...it,
          userRole: foundOrg?.userRole ?? 'vendor',
          joinedOrgAt: foundOrg?.joinedOrgAt ?? new Date().toISOString(),
        };
      }),
      tap((it: OrganizationWithUserRole) => {
        this.organizationUpdate.next(it);
      })
    );
  }
}
