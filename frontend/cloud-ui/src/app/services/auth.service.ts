import {HttpClient, HttpInterceptorFn} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {jwtDecode} from 'jwt-decode';
import {map, Observable, of, tap, throwError} from 'rxjs';
import {TokenResponse} from '../types/base';

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

  public getClaims(): {sub: string; email: string; name: string; [claim: string]: unknown} {
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
    if (token !== null) {
      return next(req.clone({headers: req.headers.set('Authorization', `Bearer ${token}`)}));
    } else {
      return throwError(() => new Error('no token'));
    }
  } else {
    return next(req);
  }
};
