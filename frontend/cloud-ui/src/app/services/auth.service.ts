import {HttpClient, HttpErrorResponse, HttpInterceptorFn} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {jwtDecode} from 'jwt-decode';
import {map, Observable, of, tap, throwError} from 'rxjs';
import {TokenResponse} from '../types/base';
import dayjs from 'dayjs';

const tokenStorageKey = 'cloud_token';

@Injectable({providedIn: 'root'})
export class AuthService {
  private readonly httpClient = inject(HttpClient);
  private readonly baseUrl = '/api/v1/auth';

  public get isAuthenticated(): boolean {
    return this.token != null;
  }

  public get token(): string | null {
    return localStorage.getItem(tokenStorageKey);
  }

  public set token(value: string | null) {
    if (value !== null) {
      localStorage.setItem(tokenStorageKey, value);
    } else {
      localStorage.removeItem(tokenStorageKey);
    }
  }

  public login(email: string, password: string): Observable<void> {
    return this.httpClient.post<TokenResponse>(`${this.baseUrl}/login`, {email, password}).pipe(
      tap((r) => (this.token = r.token)),
      map(() => undefined)
    );
  }

  public register(email: string, name: string | null | undefined, password: string): Observable<void> {
    let body: any = {email, password};
    if (name) {
      body = {...body, name};
    }
    return this.httpClient.post<void>(`${this.baseUrl}/register`, body);
  }

  public getClaims(): {sub: string; email: string; name: string; exp: string; [claim: string]: unknown} {
    if (this.token !== null) {
      return jwtDecode(this.token);
    } else {
      throw new Error('token is null');
    }
  }

  public logout(): Observable<void> {
    this.token = null;
    return of(undefined);
  }
}

export const tokenInterceptor: HttpInterceptorFn = (req, next) => {
  const auth = inject(AuthService);
  if (req.url !== '/api/v1/auth/login' && req.url !== '/api/v1/auth/register') {
    const token = auth.token;
    try {
      if (dayjs.unix(parseInt(auth.getClaims().exp)).isAfter(dayjs())) {
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
        return throwError(() => new Error('token has expired'));
      }
    } catch (cause) {
      return throwError(() => new Error('no token', {cause}));
    }
  } else {
    return next(req);
  }
};
