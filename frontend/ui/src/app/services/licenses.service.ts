import {HttpClient, HttpErrorResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {catchError, Observable, of, tap, throwError} from 'rxjs';
import {DefaultReactiveList, ReactiveList} from './cache';
import {CrudService} from './interfaces';
import {Application, ApplicationVersion} from '@glasskube/distr-sdk';
import {License} from '../types/license';

@Injectable({
  providedIn: 'root',
})
export class LicensesService implements CrudService<License> {
  private readonly licensesUrl = '/api/v1/application-licenses';
  private readonly cache: ReactiveList<License>;

  constructor(private readonly httpClient: HttpClient) {
    this.cache = new DefaultReactiveList(this.httpClient.get<License[]>(this.licensesUrl));
  }

  list(): Observable<License[]> {
    return this.cache.get();
  }

  create(license: License): Observable<License> {
    return this.httpClient.post<License>(this.licensesUrl, license).pipe(tap((it) => this.cache.save(it)));
  }

  update(license: License): Observable<License> {
    return this.httpClient
      .put<License>(`${this.licensesUrl}/${license.id}`, license)
      .pipe(tap((it) => this.cache.save(it)));
  }

  delete(license: License): Observable<void> {
    return this.httpClient
      .delete<void>(`${this.licensesUrl}/${license.id}`)
      .pipe(tap(() => this.cache.remove(license)));
  }
}
