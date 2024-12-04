import {HttpClient} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {Observable, tap} from 'rxjs';
import {Application, ApplicationVersion} from '../types/application';
import {DefaultReactiveList, ReactiveList} from './cache';
import {CrudService} from './interfaces';

@Injectable({
  providedIn: 'root',
})
export class ApplicationsService implements CrudService<Application> {
  private readonly applicationsUrl = '/api/applications';
  private readonly cache: ReactiveList<Application>;

  constructor(private readonly httpClient: HttpClient) {
    this.cache = new DefaultReactiveList(this.httpClient.get<Application[]>(this.applicationsUrl));
  }

  list(): Observable<Application[]> {
    return this.cache.get();
  }

  create(application: Application): Observable<Application> {
    return this.httpClient.post<Application>(this.applicationsUrl, application).pipe(tap((it) => this.cache.save(it)));
  }

  update(application: Application): Observable<Application> {
    return this.httpClient
      .put<Application>(`${this.applicationsUrl}/${application.id}`, application)
      .pipe(tap((it) => this.cache.save(it)));
  }

  createApplicationVersion(
    application: Application,
    applicationVersion: ApplicationVersion,
    file: File
  ): Observable<ApplicationVersion> {
    const formData = new FormData();
    formData.append('file', file);
    formData.append('applicationversion', JSON.stringify(applicationVersion));
    return this.httpClient
      .post<ApplicationVersion>(`${this.applicationsUrl}/${application.id}/versions`, formData)
      .pipe(
        tap((it) => {
          application.versions = [it, ...(application.versions || [])];
          this.cache.save(application);
        })
      );
  }

  createSample(): Observable<Application> {
    return this.httpClient
      .post<Application>(`${this.applicationsUrl}/sample`, null)
      .pipe(tap((it) => this.cache.save(it)));
  }
}
