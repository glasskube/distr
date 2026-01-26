import {HttpClient} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable, Subject, switchMap, tap} from 'rxjs';
import {ApplicationLicense} from '../types/application-license';
import {DefaultReactiveList, ReactiveList} from './cache';
import {CrudService} from './interfaces';

@Injectable({
  providedIn: 'root',
})
export class LicensesService implements CrudService<ApplicationLicense> {
  private readonly licensesUrl = '/api/v1/application-licenses';
  private readonly cache: ReactiveList<ApplicationLicense>;
  private readonly refresh$ = new Subject<void>();

  constructor(private readonly httpClient: HttpClient) {
    this.cache = new DefaultReactiveList(this.httpClient.get<ApplicationLicense[]>(this.licensesUrl));
    this.refresh$
      .pipe(
        switchMap(() => this.httpClient.get<ApplicationLicense[]>(this.licensesUrl)),
        tap((licenses) => this.cache.reset(licenses))
      )
      .subscribe();
  }

  list(applicationId?: string): Observable<ApplicationLicense[]> {
    if (applicationId) {
      return this.httpClient.get<ApplicationLicense[]>(this.licensesUrl, {params: {applicationId}});
    } else {
      return this.cache.get();
    }
  }

  refresh() {
    this.refresh$.next();
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
