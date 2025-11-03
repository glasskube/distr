import {HttpClient} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {
  CreateUpdateCustomerOrganizationRequest,
  CustomerOrganization,
  CustomerOrganizationWithUserCount,
} from '../types/customer-organization';
import {Observable} from 'rxjs';

const baseUrl = '/api/v1/customer-organizations';

@Injectable({
  providedIn: 'root',
})
export class CustomerOrganizationsService {
  private readonly httpClient = inject(HttpClient);

  public getCustomerOrganizations(): Observable<CustomerOrganizationWithUserCount[]> {
    return this.httpClient.get<CustomerOrganizationWithUserCount[]>(baseUrl);
  }

  public getCustomerOrganizationById(id: string): Observable<CustomerOrganization> {
    return this.httpClient.get<CustomerOrganization>(`${baseUrl}/${id}`);
  }

  public createCustomerOrganization(
    request: CreateUpdateCustomerOrganizationRequest
  ): Observable<CustomerOrganization> {
    return this.httpClient.post<CustomerOrganization>(baseUrl, request);
  }

  public updateCustomerOrganization(request: CustomerOrganization): Observable<CustomerOrganization> {
    return this.httpClient.put<CustomerOrganization>(`${baseUrl}/${request.id}`, request);
  }

  public deleteCustomerOrganization(id: string): Observable<void> {
    return this.httpClient.delete<void>(`${baseUrl}/${id}`);
  }
}
