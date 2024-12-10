import {HttpClient} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {Observable} from 'rxjs';
import {UserAccount, UserAccountWithRole} from '../types/user-account';

@Injectable({providedIn: 'root'})
export class UsersService {
  private readonly httpClient = inject(HttpClient);
  private readonly baseUrl = '/api/v1/user-accounts';

  public getUsers(): Observable<UserAccountWithRole[]> {
    return this.httpClient.get<UserAccountWithRole[]>(this.baseUrl);
  }
}
