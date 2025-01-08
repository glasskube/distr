import {HttpClient} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {UserAccount, UserAccountWithRole, UserRole} from '../types/user-account';

export interface CreateUserAccountRequest {
  email: string;
  name?: string;
  userRole: UserRole;
  applicationName?: string;
}

@Injectable({providedIn: 'root'})
export class UsersService {
  private readonly httpClient = inject(HttpClient);
  private readonly baseUrl = '/api/v1/user-accounts';

  public getUsers(): Observable<UserAccountWithRole[]> {
    return this.httpClient.get<UserAccountWithRole[]>(this.baseUrl);
  }

  public addUser(request: CreateUserAccountRequest): Observable<void> {
    return this.httpClient.post<void>(this.baseUrl, request);
  }

  public delete(user: UserAccount): Observable<void> {
    return this.httpClient.delete<void>(`${this.baseUrl}/${user.id}`);
  }
}
