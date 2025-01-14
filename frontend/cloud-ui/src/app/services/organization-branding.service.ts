import {HttpClient} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable, of, tap} from 'rxjs';
import {OrganizationBranding, OrganizationBrandingWithAuthor} from '../types/organization-branding';
import {CrudService} from './interfaces';
import {DeploymentTarget} from '../types/deployment-target';

@Injectable({
  providedIn: 'root',
})
export class OrganizationBrandingService {
  private readonly organizationBrandingUrl = '/api/v1/organization-branding';
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
      .put<OrganizationBranding>(
        `${this.organizationBrandingUrl}/${organizationBranding.get('id')}`,
        organizationBranding
      )
      .pipe(tap((obj) => (this.cache = obj)));
  }
}
