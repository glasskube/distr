import {HttpClient} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {UserAccountWithRole} from '@distr-sh/distr-sdk';
import posthog from 'posthog-js';
import {map, Observable, shareReplay, tap} from 'rxjs';
import {Organization, OrganizationWithUserRole} from '../types/organization';

interface ContextResponse {
  user: UserAccountWithRole;
  organization: Organization;
  availableContexts?: OrganizationWithUserRole[];
}

/**
 * ContextService should not be used directly – use UsersService and OrganizationService instead to profit
 * from getting live updates as well.
 */
@Injectable({providedIn: 'root'})
export class ContextService {
  private readonly baseUrl = '/api/v1/context';
  private readonly httpClient = inject(HttpClient);
  private readonly cache = this.httpClient.get<ContextResponse>(this.baseUrl).pipe(
    tap((ctx) => posthog.group('organization', ctx.organization.id!, {name: ctx.organization.name})),
    shareReplay(1)
  );

  public getUser(): Observable<UserAccountWithRole> {
    return this.cache.pipe(map((ctx) => ctx.user));
  }

  public getOrganization(): Observable<OrganizationWithUserRole> {
    return this.cache.pipe(
      map((ctx) => ({
        ...ctx.organization,
        userRole: ctx.user.userRole,
        joinedOrgAt: ctx.user.joinedOrgAt,
      }))
    );
  }

  public getAvailableOrganizations(): Observable<OrganizationWithUserRole[]> {
    return this.cache.pipe(map((ctx) => ctx.availableContexts ?? []));
  }
}
