import {HttpClient, HttpErrorResponse, HttpInterceptorFn} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {jwtDecode} from 'jwt-decode';
import {map, Observable, of, tap, throwError} from 'rxjs';
import dayjs from 'dayjs';
import {TokenResponse, UserRole} from '@glasskube/distr-sdk';

const tokenStorageKey = 'cloud_token';
const actionTokenStorageKey = 'distr_action_token';

export interface JWTClaims {
  sub: string;
  org: string;
  email: string;
  password_reset: boolean;
  email_verified: boolean;
  name: string;
  exp: string;
  role: UserRole;
  [claim: string]: unknown;
}

@Injectable({providedIn: 'root'})
export class AuthService {
  private readonly httpClient = inject(HttpClient);
  private readonly baseUrl = '/api/v1/auth';

  private get token(): string | null {
    return localStorage.getItem(tokenStorageKey);
  }

  private set token(value: string | null) {
    if (value !== null) {
      localStorage.setItem(tokenStorageKey, value);
    } else {
      localStorage.removeItem(tokenStorageKey);
    }
  }

  public get actionToken(): string | null {
    return sessionStorage.getItem(actionTokenStorageKey);
  }

  public set actionToken(value: string | null) {
    if (value !== null) {
      sessionStorage.setItem(actionTokenStorageKey, value);
    } else {
      sessionStorage.removeItem(actionTokenStorageKey);
    }
  }

  public hasRole(role: UserRole): boolean {
    return this.getClaims()?.role === role;
  }

  public login(email: string, password: string): Observable<void> {
    return this.httpClient.post<TokenResponse>(`${this.baseUrl}/login`, {email, password}).pipe(
      tap((r) => {
        this.token = r.token;
        this.actionToken = null;
      }),
      map(() => undefined)
    );
  }

  public resetPassword(email: string): Observable<void> {
    return this.httpClient.post<void>(`${this.baseUrl}/reset`, {email});
  }

  public register(email: string, name: string | null | undefined, password: string): Observable<void> {
    let body: any = {email, password};
    if (name) {
      body = {...body, name};
    }
    return this.httpClient.post<void>(`${this.baseUrl}/register`, body);
  }

  public getClaims(): JWTClaims | undefined {
    const {claims} = this.getTokenAndClaims();
    return claims;
  }

  public getTokenAndClaims(): {token: string | null; claims: JWTClaims | undefined} {
    const actionToken = this.actionToken;
    if (actionToken !== null) {
      console.log('actionToken', actionToken);
      try {
        return {token: actionToken, claims: jwtDecode(actionToken)};
      } catch (e) {
        console.error(e);
      }
    } else {
      const token = this.token;
      console.log('token', token);
      if (token !== null) {
        try {
          return {token, claims: jwtDecode(token)};
        } catch (e) {
          console.error(e);
        }
      }
    }
    return {token: null, claims: undefined};
  }

  public logout(): Observable<void> {
    this.token = null;
    this.actionToken = null;
    return of(undefined);
  }
}

export const tokenInterceptor: HttpInterceptorFn = (req, next) => {
  const auth = inject(AuthService);
  if (!req.url.startsWith('/api/v1/auth/')) {
    const {token, claims} = auth.getTokenAndClaims();
    try {
      if (claims && dayjs.unix(parseInt(claims.exp)).isAfter(dayjs())) {
        return next(req.clone({headers: req.headers.set('Authorization', `Bearer ${token}`)})).pipe(
          tap({
            error: (e) => {
              if (e instanceof HttpErrorResponse && e.status == 401) {
                auth.logout();
                location.reload();
              }
            },
          })
        );
      } else {
        auth.logout();
        location.reload();
        return throwError(() => new Error('no token or token has expired'));
      }
    } catch (cause) {
      return throwError(() => new Error('no token', {cause}));
    }
  } else {
    return next(req);
  }
};
