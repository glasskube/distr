import {HttpClient} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {map, merge, Observable, of, shareReplay, Subject, switchMap, tap} from 'rxjs';
import {UserAccountWithRole, UserRole} from '@glasskube/distr-sdk';
import {ReactiveList} from './cache';
import {AuthService} from './auth.service';
import {ContextService} from './context.service';
import {Organization} from '../types/organization';

export interface CreateUserAccountRequest {
  email: string;
  name?: string;
  userRole: UserRole;
}

export interface UserAccountInvitationResponse {
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
  private readonly contextService = inject(ContextService);
  private readonly cache = new UserAccountsReactiveList(this.httpClient.get<UserAccountWithRole[]>(this.baseUrl));

  private readonly selfUpdate = new Subject<UserAccountWithRole>();
  private readonly self$ = merge(this.selfUpdate.asObservable(), this.contextService.getUser()).pipe(shareReplay(1));

  public get(): Observable<UserAccountWithRole> {
    return this.self$;
  }

  public getUsers(): Observable<UserAccountWithRole[]> {
    return this.cache.get();
  }

  public getUserStatus(): Observable<{active: boolean}> {
    return this.httpClient.get<{active: boolean}>(`${this.baseUrl}/status`);
  }

  public addUser(request: CreateUserAccountRequest): Observable<UserAccountInvitationResponse> {
    return this.httpClient
      .post<UserAccountInvitationResponse>(this.baseUrl, request)
      .pipe(tap((it) => this.cache.save(it.user)));
  }

  public resendInvitation(user: UserAccountWithRole): Observable<UserAccountInvitationResponse> {
    return this.httpClient
      .post<UserAccountInvitationResponse>(`${this.baseUrl}/${user.id}/invite`, undefined)
      .pipe(tap((it) => this.cache.save(it.user)));
  }

  public delete(user: UserAccountWithRole): Observable<void> {
    return this.httpClient.delete<void>(`${this.baseUrl}/${user.id}`).pipe(tap(() => this.cache.remove(user)));
  }

  public patchImage(userId: string, imageId: string) {
    return this.httpClient.patch<UserAccountWithRole>(`${this.baseUrl}/${userId}/image`, {imageId}).pipe(
      tap((it) => {
        this.cache.save(it);
        this.selfUpdate.next(it);
      })
    );
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
}
