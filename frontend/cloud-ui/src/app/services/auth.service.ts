import {HttpClient, HttpInterceptorFn} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {EMPTY, ignoreElements, lastValueFrom, map, mapTo, Observable, of, tap, throwError} from 'rxjs';

const tokenStorageKey = 'cloud_token';

@Injectable({providedIn: 'root'})
export class AuthService {
  private readonly httpClient = inject(HttpClient);
  private readonly baseUrl = '/api/auth';

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
    return this.httpClient.post<{token: string}>(`${this.baseUrl}/login`, {email, password}).pipe(
      tap((r) => (this.token = r.token)),
      map(() => undefined)
    );
  }

  public logout(): Observable<void> {
    this.token = null;
    return of(undefined);
  }

  public httpInterceptor(): HttpInterceptorFn {
    return (req, next) => {
      if (req.url !== '/api/auth/login') {
        const token = this.token;
        if (token !== null) {
          return next(req.clone({headers: req.headers.set('Authentication', `Bearer ${this.token}`)}));
        } else {
          return throwError(() => new Error('no token'));
        }
      } else {
        return next(req);
      }
    };
  }
}

export const tokenInterceptor: HttpInterceptorFn = (req, next) => {
  const auth = inject(AuthService);
  if (req.url !== '/api/auth/login') {
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
