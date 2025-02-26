import {Injectable} from '@angular/core';
import {EMPTY, first, map, Observable, of} from 'rxjs';
import {BaseModel, Named, UserAccount} from '@glasskube/distr-sdk';
import {Artifact, ArtifactTag} from './artifacts.service';
import {artifactsMock, generateUUIDv4} from './artifacts-mock';
import {CrudService} from './interfaces';
import {DefaultReactiveList} from './cache';

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
  private readonly cache = new DefaultReactiveList(
    of([
      {
        id: 'b135b6b2-ebc9-4c13-a2c1-7eaa79455955',
        name: 'distr',
        createdAt: '2025-02-25T09:25:21Z',
        artifactId: artifactsMock[0].id,
        artifact: artifactsMock[0],
      } as ArtifactLicense,
      {
        id: '49638b03-4644-4221-81df-be8981622c74',
        name: 'distr-docker-agent',
        createdAt: '2025-02-25T09:25:21Z',
        artifactId: artifactsMock[1].id,
        artifact: artifactsMock[1],
        artifactTags: [artifactsMock[1].tags[0]],
      } as ArtifactLicense,
    ])
  );

  public list(): Observable<ArtifactLicense[]> {
    return this.cache.get();
  }

  create(request: ArtifactLicense): Observable<ArtifactLicense> {
    request.id = generateUUIDv4();
    request.artifact = artifactsMock.find((a) => a.id === request.artifactId);
    this.cache.save(request);
    return of(request);
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
      map((oldLicense) => {
        const newLicense = {
          ...oldLicense,
          ...request,
          artifact: artifactsMock.find((a) => a.id === request.artifactId),
        };
        this.cache.save(newLicense);
        return newLicense;
      })
    );
  }
}
