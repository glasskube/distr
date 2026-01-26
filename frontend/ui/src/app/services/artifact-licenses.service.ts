import {HttpClient} from '@angular/common/http';
import {Injectable} from '@angular/core';
import {BaseModel, Named} from '@distr-sh/distr-sdk';
import {Observable, Subject, switchMap, tap} from 'rxjs';
import {DefaultReactiveList, ReactiveList} from './cache';
import {CrudService} from './interfaces';

export interface ArtifactLicenseSelection {
  artifactId: string;
  versionIds?: string[];
}

export interface ArtifactLicense extends BaseModel, Named {
  expiresAt?: Date;
  artifacts?: ArtifactLicenseSelection[];
  customerOrganizationId?: string;
}

@Injectable({providedIn: 'root'})
export class ArtifactLicensesService implements CrudService<ArtifactLicense> {
  private readonly cache: ReactiveList<ArtifactLicense>;
  private readonly artifactLicensesUrl = '/api/v1/artifact-licenses';
  private readonly refresh$ = new Subject<void>();

  constructor(private readonly http: HttpClient) {
    this.cache = new DefaultReactiveList(this.http.get<ArtifactLicense[]>(this.artifactLicensesUrl));
    this.refresh$
      .pipe(
        switchMap(() => this.http.get<ArtifactLicense[]>(this.artifactLicensesUrl)),
        tap((licenses) => this.cache.reset(licenses))
      )
      .subscribe();
  }

  public list(): Observable<ArtifactLicense[]> {
    return this.cache.get();
  }

  refresh() {
    this.refresh$.next();
  }

  create(request: ArtifactLicense): Observable<ArtifactLicense> {
    return this.http.post<ArtifactLicense>(this.artifactLicensesUrl, request).pipe(tap((l) => this.cache.save(l)));
  }

  delete(request: ArtifactLicense): Observable<void> {
    return this.http
      .delete<void>(`${this.artifactLicensesUrl}/${request.id}`)
      .pipe(tap(() => this.cache.remove(request)));
  }

  update(request: ArtifactLicense): Observable<ArtifactLicense> {
    return this.http
      .put<ArtifactLicense>(`${this.artifactLicensesUrl}/${request.id}`, request)
      .pipe(tap((l) => this.cache.save(l)));
  }
}
