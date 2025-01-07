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
  private readonly applicationsUrl = '/api/v1/applications';
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

  delete(application: Application): Observable<void> {
    return this.httpClient
      .delete<void>(`${this.applicationsUrl}/${application.id}`)
      .pipe(tap(() => this.cache.remove(application)));
  }

  createApplicationVersionForDocker(
    application: Application,
    applicationVersion: ApplicationVersion,
    file: File
  ): Observable<ApplicationVersion> {
    const formData = new FormData();
    formData.append('composefile', file);
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

  createApplicationVersionForKubernetes(
    application: Application,
    applicationVersion: ApplicationVersion,
    valuesFile: File | null,
    templateFile: File | null
  ): Observable<ApplicationVersion> {
    const formData = new FormData();
    if (valuesFile) {
      formData.append('valuesfile', valuesFile);
    }
    if (templateFile) {
      formData.append('templatefile', templateFile);
    }
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
