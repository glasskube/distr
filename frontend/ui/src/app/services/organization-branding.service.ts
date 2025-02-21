import {HttpClient} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable, of, tap} from 'rxjs';
import {OrganizationBranding} from '@glasskube/distr-sdk';

@Injectable({
  providedIn: 'root',
})
export class OrganizationBrandingService {
  private readonly organizationBrandingUrl = '/api/v1/organization/branding';
  private cache?: OrganizationBranding;

  constructor(private readonly httpClient: HttpClient) {}

  get(): Observable<OrganizationBranding> {
    if (this.cache) {
      return of(this.cache);
    }
    return this.httpClient
      .get<OrganizationBranding>(this.organizationBrandingUrl)
      .pipe(tap((branding) => (this.cache = branding)));
  }

  create(organizationBranding: FormData): Observable<OrganizationBranding> {
    return this.httpClient
      .post<OrganizationBranding>(this.organizationBrandingUrl, organizationBranding)
      .pipe(tap((obj) => (this.cache = obj)));
  }

  update(organizationBranding: FormData): Observable<OrganizationBranding> {
    return this.httpClient
      .put<OrganizationBranding>(this.organizationBrandingUrl, organizationBranding)
      .pipe(tap((obj) => (this.cache = obj)));
  }
}
