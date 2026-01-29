import {HttpClient} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {UserAccount} from '@distr-sh/distr-sdk';
import {Observable, tap} from 'rxjs';
import {ContextService} from './context.service';

export interface UpdateUserAccountRequest {
  name?: string;
  password?: string;
  emailVerified?: boolean;
  imageId?: string;
}

export interface MFASetupData {
  qrCodeUrl: string;
  secret: string;
}

@Injectable({providedIn: 'root'})
export class SettingsService {
  private readonly httpClient = inject(HttpClient);
  private readonly ctx = inject(ContextService);
  private readonly baseUrl = '/api/v1/settings';

  public updateUserSettings(request: UpdateUserAccountRequest): Observable<UserAccount> {
    return this.httpClient.post<UserAccount>(`${this.baseUrl}/user`, request).pipe(tap(() => this.ctx.reload()));
  }

  public requestEmailVerification(email?: string) {
    if (email) {
      return this.httpClient.post<void>(`${this.baseUrl}/user/email`, {email});
    } else {
      return this.httpClient.post<void>(`${this.baseUrl}/verify/request`, undefined);
    }
  }

  public confirmEmailVerification() {
    return this.httpClient.post<void>(`${this.baseUrl}/verify/confirm`, undefined);
  }

  public startMFASetup() {
    return this.httpClient.post<MFASetupData>(`${this.baseUrl}/mfa/setup`, undefined);
  }

  public enableMFA(code: string) {
    return this.httpClient.post<void>(`${this.baseUrl}/mfa/enable`, {code}).pipe(tap(() => this.ctx.reload()));
  }

  public disableMFA(password: string) {
    return this.httpClient.post<void>(`${this.baseUrl}/mfa/disable`, {password}).pipe(tap(() => this.ctx.reload()));
  }
}
