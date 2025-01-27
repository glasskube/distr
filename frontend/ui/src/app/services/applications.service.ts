import {HttpClient, HttpErrorResponse} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {catchError, Observable, of, tap, throwError} from 'rxjs';
import {DefaultReactiveList, ReactiveList} from './cache';
import {CrudService} from './interfaces';
import {Application, ApplicationVersion} from '@glasskube/distr-sdk';

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

  getTemplateFile(applicationId: string, versionId: string): Observable<string | null> {
    return this.getFile(applicationId, versionId, 'template-file');
  }

  getValuesFile(applicationId: string, versionId: string): Observable<string | null> {
    return this.getFile(applicationId, versionId, 'values-file');
  }

  getComposeFile(applicationId: string, versionId: string): Observable<string | null> {
    return this.getFile(applicationId, versionId, 'compose-file');
  }

  private getFile(applicationId: string, versionId: string, file: string): Observable<string | null> {
    return this.httpClient
      .get(`${this.applicationsUrl}/${applicationId}/versions/${versionId}/${file}`, {responseType: 'text'})
      .pipe(
        catchError((e) => {
          if (e instanceof HttpErrorResponse && e.status == 404) {
            return of(null);
          } else {
            return throwError(() => e);
          }
        })
      );
  }

  createApplicationVersionForDocker(
    application: Application,
    applicationVersion: ApplicationVersion,
    compose: string
  ): Observable<ApplicationVersion> {
    const formData = new FormData();
    formData.append('composefile', new Blob([compose], {type: 'application/yaml'}));
    return this.doCreateVersion(application, applicationVersion, formData);
  }

  createApplicationVersionForKubernetes(
    application: Application,
    applicationVersion: ApplicationVersion,
    baseValues?: string | null,
    template?: string | null
  ): Observable<ApplicationVersion> {
    const formData = new FormData();
    if (baseValues) {
      formData.append('valuesfile', new Blob([baseValues], {type: 'application/yaml'}));
    }
    if (template) {
      formData.append('templatefile', new Blob([template], {type: 'application/yaml'}));
    }
    return this.doCreateVersion(application, applicationVersion, formData);
  }

  private doCreateVersion(application: Application, applicationVersion: ApplicationVersion, formData: FormData) {
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
