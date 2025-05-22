import {HttpClient} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {map, Observable, of, switchMap, tap} from 'rxjs';
import {UserAccountWithRole, UserRole} from '@glasskube/distr-sdk';
import {ReactiveList} from './cache';
import {AuthService} from './auth.service';

export interface CreateUserAccountRequest {
  email: string;
  name?: string;
  userRole: UserRole;
}

export interface CreateUserAccountResponse {
  user: UserAccountWithRole;
  inviteUrl: string;
}

class UserAccountsReactiveList extends ReactiveList<UserAccountWithRole> {
  protected override identify = (u: UserAccountWithRole) => u.id;
  protected override sortAttr = (u: UserAccountWithRole) => u.name ?? u.email;
}

@Injectable({providedIn: 'root'})
export class UsersService {
  private readonly baseUrl = '/api/v1/user-accounts';
  private readonly httpClient = inject(HttpClient);
  private readonly cache = new UserAccountsReactiveList(this.httpClient.get<UserAccountWithRole[]>(this.baseUrl));
  private readonly auth = inject(AuthService);

  public getUsers(): Observable<UserAccountWithRole[]> {
    if (this.auth.hasRole('customer')) {
      const claims = this.auth.getClaims();
      if (claims) {
        return of([
          {
            id: claims.sub,
            email: claims.email,
            name: claims.name,
            userRole: 'customer',
            imageUrl: claims.image_url,
          },
        ]);
      }
      return of([]);
    }
    return this.cache.get();
  }

  public getUserStatus(): Observable<{active: boolean}> {
    return this.httpClient.get<{active: boolean}>(`${this.baseUrl}/status`);
  }

  public addUser(request: CreateUserAccountRequest): Observable<CreateUserAccountResponse> {
    return this.httpClient
      .post<CreateUserAccountResponse>(this.baseUrl, request)
      .pipe(tap((it) => this.cache.save(it.user)));
  }

  public delete(user: UserAccountWithRole): Observable<void> {
    return this.httpClient.delete<void>(`${this.baseUrl}/${user.id}`).pipe(tap(() => this.cache.remove(user)));
  }

  public patchImage(userId: string, imageId: string) {
    return this.httpClient
      .patch<UserAccountWithRole>(`${this.baseUrl}/${userId}/image`, {imageId})
      .pipe(tap((it) => this.cache.save(it)));
  }

  public getUser(id: string): Observable<UserAccountWithRole> {
    return this.getUsers().pipe(
      map((users) => users.find((u) => u.id === id)),
      map((u) => {
        if (!u) {
          throw 'user not found';
        }
        return u;
      })
    );
  }

  public getUserByEmail(email: string): Observable<UserAccountWithRole> {
    return this.getUsers().pipe(
      map((users) => users.find((u) => u.email === email)),
      map((u) => {
        if (!u) {
          throw 'user not found';
        }
        return u;
      })
    );
  }
}
