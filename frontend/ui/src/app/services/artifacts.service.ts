import {inject, Injectable} from '@angular/core';
import {combineLatestWith, first, from, lastValueFrom, map, Observable, of, switchMap, tap, withLatestFrom} from 'rxjs';
import {digestMessage} from '../../util/crypto';
import {AuthService} from './auth.service';
import {HttpClient} from '@angular/common/http';
import {DefaultReactiveList, ReactiveList} from './cache';
import {Application} from '../../../../../sdk/js/src';

export interface HasDownloads {
  downloadsTotal?: number;
  downloadedByCount?: number;
  downloadedByUsers?: string[];
}

export interface ArtifactUser {
  id: string;
  avatarUrl: string;
}

export interface VulnerabilitySeverity {
  type: 'CVSS_V2' | 'CVSS_V3' | 'CVSS_V4';
  score: string;
}

/**
 * From https://ossf.github.io/osv-schema/
 *
 * Severity calculator: https://www.first.org/cvss/calculator/4.0
 */
export interface Vulnerability {
  id: string;
  severity: VulnerabilitySeverity[];
}

export interface BaseArtifact {
  id: string;
  name: string;
}

export interface BaseArtifactVersion {
  id: string;
  name: string;
}

export interface Artifact extends BaseArtifact, HasDownloads {}

export interface TaggedArtifactVersion extends HasDownloads {
  id: string;
  digest: string;
  sbom?: string;
  createdAt: string;
  size: number;
  tags: {name: string; downloads: HasDownloads}[];
  vulnerabilities: Vulnerability[];
  lastScannedAt?: string;
}

export interface ArtifactWithTags extends Artifact {
  versions: TaggedArtifactVersion[];
}

@Injectable({providedIn: 'root'})
export class ArtifactsService {
  private readonly auth = inject(AuthService);
  private readonly cache: ReactiveList<ArtifactWithTags>;
  private readonly artifactsUrl = '/api/v1/artifacts';

  constructor(private readonly http: HttpClient) {
    this.cache = new DefaultReactiveList(this.http.get<ArtifactWithTags[]>(this.artifactsUrl));
  }

  private async getDownloadedByUsers(self: boolean = true, count = 3): Promise<ArtifactUser[]> {
    if (this.auth.hasRole('vendor')) {
      if (count === 1) {
        return [{id: '4f21317b-61d5-44a8-a431-c220f3fd010f', avatarUrl: '/placeholders/company-4.jpg'}];
      }

      return [
        {id: '4f21317b-61d5-44a8-a431-c220f3fd010f', avatarUrl: '/placeholders/company-1.jpg'},
        {id: '45560805-6900-4160-ba32-1d9f09bafff6', avatarUrl: '/placeholders/company-2.jpg'},
        {id: 'e3605a1d-4a91-4cba-9137-574f24d07c72', avatarUrl: '/placeholders/company-3.jpg'},
      ];
    }

    if (self) {
      const email = this.auth.getClaims()?.email;
      if (email) {
        return [
          {
            id: this.auth.getClaims()?.sub ?? '',
            avatarUrl: `https://www.gravatar.com/avatar/${await digestMessage(email)}`,
          },
        ];
      }
    }

    return [];
  }

  public list(): Observable<ArtifactWithTags[]> {
    return this.cache.get();
  }

  public getByIdAndCache(id: string): Observable<ArtifactWithTags | undefined> {
    return this.list().pipe(
      map((ls) => ls.find((a) => a.id === id)),
      switchMap((existing) => {
        if ((existing?.versions ?? []).length > 0) {
          return of(existing);
        } else if (existing) {
          return this.http.get<ArtifactWithTags>(`${this.artifactsUrl}/${id}`).pipe(tap((a) => this.cache.save(a)));
        } else {
          return of(undefined);
        }
      })
    );
  }
}
