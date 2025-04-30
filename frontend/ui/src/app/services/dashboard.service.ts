import {HttpClient} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {UserAccountWithRole} from '@glasskube/distr-sdk';
import {ArtifactWithTags} from './artifacts.service';

export interface DashboardArtifact {
  artifact: ArtifactWithTags;
  latestPulledVersion: string;
}

export interface ArtifactsByCustomer {
  customer: UserAccountWithRole;
  artifacts?: DashboardArtifact[];
}

@Injectable({providedIn: 'root'})
export class DashboardService {
  private readonly httpClient = inject(HttpClient);
  private readonly baseUrl = '/api/v1/dashboard';

  public getArtifactsByCustomer(): Observable<ArtifactsByCustomer[]> {
    return this.httpClient.get<ArtifactsByCustomer[]>(`${this.baseUrl}/artifacts-by-customer`);
  }
}
