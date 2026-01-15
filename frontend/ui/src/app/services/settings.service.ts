import {HttpClient} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {UserAccount} from '@distr-sh/distr-sdk';
import {Observable} from 'rxjs';

@Injectable({providedIn: 'root'})
export class SettingsService {
  private readonly httpClient = inject(HttpClient);
  private readonly baseUrl = '/api/v1/settings';

  public updateUserSettings(request: {
    name?: string;
    password?: string;
    emailVerified?: boolean;
  }): Observable<UserAccount> {
    return this.httpClient.post<UserAccount>(`${this.baseUrl}/user`, request);
  }

  public requestEmailVerification() {
    return this.httpClient.post<void>(`${this.baseUrl}/verify/request`, undefined);
  }

  public confirmEmailVerification() {
    return this.httpClient.post<void>(`${this.baseUrl}/verify/confirm`, undefined);
  }
}
