import {inject, Injectable} from '@angular/core';
import {combineLatestWith, EMPTY, first, map, Observable, tap} from 'rxjs';
import {BaseModel, Named, UserAccount} from '@glasskube/distr-sdk';
import {Artifact, ArtifactsService, ArtifactTag} from './artifacts.service';
import {CrudService} from './interfaces';
import {DefaultReactiveList} from './cache';
import {UsersService} from './users.service';

export interface ArtifactLicense extends BaseModel, Named {
  expiresAt?: Date;
  artifactId?: string;
  artifact?: Artifact;
  artifactTags?: ArtifactTag[];
  ownerUserAccountId?: string;
  owner?: UserAccount;
}

@Injectable({providedIn: 'root'})
export class ArtifactLicensesService implements CrudService<ArtifactLicense> {
  private readonly artifactsService = inject(ArtifactsService);
  private readonly usersService = inject(UsersService);
  private readonly cache = new DefaultReactiveList<ArtifactLicense>(
    this.artifactsService.list().pipe(
      first(),
      map((mockArtifacts) => {
        return [
          {
            id: 'b135b6b2-ebc9-4c13-a2c1-7eaa79455955',
            name: 'distr',
            createdAt: '2025-02-25T09:25:21Z',
            artifactId: mockArtifacts[0].id,
            artifact: mockArtifacts[0],
          } as ArtifactLicense,
          {
            id: '49638b03-4644-4221-81df-be8981622c74',
            name: 'distr-docker-agent',
            createdAt: '2025-02-25T09:25:21Z',
            artifactId: mockArtifacts[1].id,
            artifact: mockArtifacts[1],
            artifactTags: [mockArtifacts[1].tags[0]],
          } as ArtifactLicense,
        ];
      })
    )
  );

  public list(): Observable<ArtifactLicense[]> {
    return this.cache.get();
  }

  create(request: ArtifactLicense): Observable<ArtifactLicense> {
    return this.artifactsService.get(request.artifactId!).pipe(
      combineLatestWith(
        this.usersService.getUsers().pipe(map((users) => users.find((u) => u.id === request.ownerUserAccountId)))
      ),
      map(([artifact, owner]) => {
        return {
          ...request,
          id: generateUUIDv4(),
          artifact,
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
        this.artifactsService.get(request.artifactId!),
        this.usersService.getUsers().pipe(map((users) => users.find((u) => u.id === request.ownerUserAccountId)))
      ),
      map(([oldLicense, artifact, owner]) => {
        const newLicense = {
          ...oldLicense,
          ...request,
          artifact,
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
