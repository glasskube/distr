import {Injectable} from '@angular/core';
import {map, Observable, of, switchMap, tap} from 'rxjs';
import {HttpClient} from '@angular/common/http';
import {DefaultReactiveList, ReactiveList} from './cache';

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
  imageUrl?: string;
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
  imageUrl?: string;
}

export interface ArtifactWithTags extends Artifact {
  versions?: TaggedArtifactVersion[];
}

@Injectable({providedIn: 'root'})
export class ArtifactsService {
  private readonly cache: ReactiveList<ArtifactWithTags>;
  private readonly artifactsUrl = '/api/v1/artifacts';

  constructor(private readonly http: HttpClient) {
    this.cache = new DefaultReactiveList(this.http.get<ArtifactWithTags[]>(this.artifactsUrl));
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

  public patchImage(artifactsId: string, imageId: string) {
    return this.http
      .patch<ArtifactWithTags>(`${this.artifactsUrl}/${artifactsId}/image`, {imageId})
      .pipe(tap((it) => this.cache.save(it)));
  }
}
