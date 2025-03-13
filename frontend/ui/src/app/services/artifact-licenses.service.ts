import {inject, Injectable} from '@angular/core';
import {combineLatestWith, EMPTY, first, map, Observable, tap} from 'rxjs';
import {BaseModel, Named, UserAccount} from '@glasskube/distr-sdk';
import {Artifact, ArtifactsService, ArtifactWithTags, TaggedArtifactVersion} from './artifacts.service';
import {CrudService} from './interfaces';
import {DefaultReactiveList, ReactiveList} from './cache';
import {UsersService} from './users.service';
import {HttpClient} from '@angular/common/http';

export interface ArtifactLicenseSelection {
  artifact: Artifact;
  tags?: TaggedArtifactVersion[];
}

export interface ArtifactLicense extends BaseModel, Named {
  expiresAt?: Date;
  artifacts?: ArtifactLicenseSelection[];
  ownerUserAccountId?: string;
  owner?: UserAccount;
}

@Injectable({providedIn: 'root'})
export class ArtifactLicensesService implements CrudService<ArtifactLicense> {
  private readonly usersService = inject(UsersService);
  private readonly cache: ReactiveList<ArtifactLicense>;
  private readonly artifactLicensesUrl = '/api/v1/artifact-licenses';

  constructor(private readonly http: HttpClient) {
    this.cache = new DefaultReactiveList(this.http.get<ArtifactLicense[]>(this.artifactLicensesUrl));
  }

  public list(): Observable<ArtifactLicense[]> {
    return this.cache.get();
  }

  create(request: ArtifactLicense): Observable<ArtifactLicense> {
    return this.usersService
      .getUsers()
      .pipe(map((users) => users.find((u) => u.id === request.ownerUserAccountId)))
      .pipe(
        map((owner) => {
          return {
            ...request,
            id: generateUUIDv4(),
            owner,
          } as ArtifactLicense;
        }),
        tap((t) => this.cache.save(t))
      );
  }

  delete(request: ArtifactLicense): Observable<void> {
    this.cache.remove(request);
    return EMPTY;
  }

  update(request: ArtifactLicense): Observable<ArtifactLicense> {
    return this.list().pipe(
      first(),
      map((licenses) => {
        return licenses.find((l) => l.id === request.id);
      }),
      combineLatestWith(
        this.usersService.getUsers().pipe(map((users) => users.find((u) => u.id === request.ownerUserAccountId)))
      ),
      map(([oldLicense, owner]) => {
        const newLicense = {
          ...oldLicense,
          ...request,
          owner,
        };
        this.cache.save(newLicense);
        return newLicense;
      })
    );
  }
}

function generateUUIDv4(): string {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    const v = c === 'x' ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}
