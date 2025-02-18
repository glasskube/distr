import {HttpClient, HttpErrorResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {catchError, Observable, of, tap, throwError} from 'rxjs';
import {DefaultReactiveList, ReactiveList} from './cache';
import {CrudService} from './interfaces';
import {Application, ApplicationVersion} from '@glasskube/distr-sdk';
import {ApplicationLicense} from '../types/application-license';

@Injectable({
  providedIn: 'root',
})
export class LicensesService implements CrudService<ApplicationLicense> {
  private readonly licensesUrl = '/api/v1/application-licenses';
  private readonly cache: ReactiveList<ApplicationLicense>;

  constructor(private readonly httpClient: HttpClient) {
    this.cache = new DefaultReactiveList(this.httpClient.get<ApplicationLicense[]>(this.licensesUrl));
  }

  list(): Observable<ApplicationLicense[]> {
    return this.cache.get();
  }

  create(license: ApplicationLicense): Observable<ApplicationLicense> {
    return this.httpClient.post<ApplicationLicense>(this.licensesUrl, license).pipe(tap((it) => this.cache.save(it)));
  }

  update(license: ApplicationLicense): Observable<ApplicationLicense> {
    return this.httpClient
      .put<ApplicationLicense>(`${this.licensesUrl}/${license.id}`, license)
      .pipe(tap((it) => this.cache.save(it)));
  }

  delete(license: ApplicationLicense): Observable<void> {
    return this.httpClient
      .delete<void>(`${this.licensesUrl}/${license.id}`)
      .pipe(tap(() => this.cache.remove(license)));
  }
}
