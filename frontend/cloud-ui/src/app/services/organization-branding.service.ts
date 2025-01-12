import {HttpClient} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {OrganizationBranding, OrganizationBrandingWithAuthor} from '../types/organization-branding';

@Injectable({
  providedIn: 'root',
})
export class OrganizationBrandingService {
  private readonly organizationBrandingUrl = '/api/v1/organization-branding';

  constructor(private readonly httpClient: HttpClient) {}

  get(): Observable<OrganizationBrandingWithAuthor> {
    return this.httpClient.get<OrganizationBrandingWithAuthor>(this.organizationBrandingUrl);
  }

  create(organizationBranding: FormData): Observable<OrganizationBranding> {
    return this.httpClient.post<OrganizationBranding>(this.organizationBrandingUrl, organizationBranding);
  }

  update(organizationBranding: FormData): Observable<OrganizationBranding> {
    return this.httpClient.put<OrganizationBranding>(
      `${this.organizationBrandingUrl}/${organizationBranding.get('id')}`,
      organizationBranding
    );
  }
}
