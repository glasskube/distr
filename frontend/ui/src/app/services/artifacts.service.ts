import {Injectable} from '@angular/core';
import {Observable, of, throwError} from 'rxjs';
import {artifactsMock} from './artifacts-mock';

export interface HasDownloads {
  downloadsTotal: number;
  downloadedByCount: number;
  downloadedByUsers: {avatarUrl: string}[];
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

export interface Artifact extends HasDownloads {
  id: string;
  name: string;
}

export interface ArtifactTag extends HasDownloads {
  id: string;
  hash: string;
  sbom?: string;
  createdAt: string;
  labels: {name: string}[];
  vulnerabilities: Vulnerability[];
}

export interface ArtifactWithTags extends Artifact {
  tags: ArtifactTag[];
}

@Injectable({providedIn: 'root'})
export class ArtifactsService {
  private artifacts: ArtifactWithTags[] = [...artifactsMock]

  public list(): Observable<ArtifactWithTags[]> {
    return of(this.artifacts);
  }

  public get(id: string): Observable<ArtifactWithTags> {
    const artifact = this.artifacts.find((it) => it.id === id);
    if (artifact !== undefined) {
      return of(artifact);
    } else {
      return throwError(() => new Error('not found'));
    }
  }
}
