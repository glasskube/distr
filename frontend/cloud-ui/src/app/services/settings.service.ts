import {HttpClient} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {UserAccount} from '../types/user-account';

@Injectable({providedIn: 'root'})
export class SettingsService {
  private readonly httpClient = inject(HttpClient);
  private readonly baseUrl = '/api/v1/settings';

  public updateUserSettings(request: {name?: string; password?: string, emailVerified?: boolean}): Observable<UserAccount> {
    return this.httpClient.post<UserAccount>(`${this.baseUrl}/user`, request);
  }
}
