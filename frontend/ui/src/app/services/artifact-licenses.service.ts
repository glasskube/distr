import {inject, Injectable} from '@angular/core';
import {combineLatestWith, EMPTY, first, map, Observable, tap} from 'rxjs';
import {BaseModel, Named, UserAccount} from '@glasskube/distr-sdk';
import {Artifact, ArtifactsService, ArtifactWithTags, TaggedArtifactVersion} from './artifacts.service';
import {CrudService} from './interfaces';
import {DefaultReactiveList, ReactiveList} from './cache';
import {UsersService} from './users.service';
import {HttpClient} from '@angular/common/http';

export interface ArtifactLicenseSelection {
  artifactId: string;
  versionIds?: string[];
}

export interface ArtifactLicense extends BaseModel, Named {
  expiresAt?: Date;
  artifacts?: ArtifactLicenseSelection[];
  ownerUserAccountId?: string;
}

@Injectable({providedIn: 'root'})
export class ArtifactLicensesService implements CrudService<ArtifactLicense> {
  private readonly cache: ReactiveList<ArtifactLicense>;
  private readonly artifactLicensesUrl = '/api/v1/artifact-licenses';

  constructor(private readonly http: HttpClient) {
    this.cache = new DefaultReactiveList(this.http.get<ArtifactLicense[]>(this.artifactLicensesUrl));
  }

  public list(): Observable<ArtifactLicense[]> {
    return this.cache.get();
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
