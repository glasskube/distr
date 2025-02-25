import {Injectable} from '@angular/core';
import {Observable, of, throwError} from 'rxjs';

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
  private artifacts: ArtifactWithTags[] = [
    {
      id: '86d0f4c2-650c-480f-b875-ae2857c9753f',
      name: 'distr',
      downloadsTotal: 782,
      downloadedByCount: 72,
      downloadedByUsers: [
        {avatarUrl: '/placeholders/company-1.jpg'},
        {avatarUrl: '/placeholders/company-2.jpg'},
        {avatarUrl: '/placeholders/company-3.jpg'},
      ],
      tags: [
        {
          hash: 'sha265:78f8664cbfbec1c378f8c2af68f6fcbb1ce3faf1388c9d0b70533152b1415e98',
          sbom: 'aaaaaaaaaaaaa',
          createdAt: '2025-02-25T09:25:21Z',
          downloadsTotal: 345,
          downloadedByCount: 12,
          downloadedByUsers: [
            {avatarUrl: '/placeholders/company-1.jpg'},
            {avatarUrl: '/placeholders/company-2.jpg'},
            {avatarUrl: '/placeholders/company-3.jpg'},
          ],
          labels: [{name: 'latest'}, {name: '1.2.1'}],
          vulnerabilities: [],
        },
        {
          hash: 'sha265:28b7a85914586d15a531566443b6d5ea6d11ad38b1e75fa753385f03b0a0a57f',
          createdAt: '2025-02-25T09:25:21Z',
          downloadsTotal: 124,
          downloadedByCount: 79,
          downloadedByUsers: [
            {avatarUrl: '/placeholders/company-1.jpg'},
            {avatarUrl: '/placeholders/company-2.jpg'},
            {avatarUrl: '/placeholders/company-3.jpg'},
          ],
          labels: [{name: '1.1.6'}],
          vulnerabilities: [
            {
              id: 'GHSA-vp9c-fpxx-744v',
              severity: [{type: 'CVSS_V4', score: 'CVSS:4.0/AV:N/AC:L/AT:N/PR:N/UI:N/VC:H/VI:H/VA:N/SC:N/SI:N/SA:N'}],
            },
            {
              id: 'CVE-2025-375',
              severity: [{type: 'CVSS_V4', score: 'CVSS:4.0/AV:N/AC:L/AT:N/PR:N/UI:N/VC:L/VI:L/VA:N/SC:N/SI:N/SA:N'}],
            },
            {
              id: 'GO-2025-2345',
              severity: [{type: 'CVSS_V4', score: 'CVSS:4.0/AV:N/AC:L/AT:N/PR:N/UI:N/VC:L/VI:H/VA:N/SC:N/SI:N/SA:N'}],
            },
            {
              id: 'CVE-2024-34854',
              severity: [{type: 'CVSS_V4', score: 'CVSS:4.0/AV:N/AC:L/AT:P/PR:H/UI:A/VC:N/VI:N/VA:L/SC:N/SI:N/SA:N'}],
            },
          ],
        },
      ],
    },
    {
      id: '6c988429-cef7-45ce-9ef5-a4af55cfc8a2',
      name: 'distr/docker-agent',
      downloadsTotal: 1234,
      downloadedByCount: 759,
      downloadedByUsers: [
        {avatarUrl: '/placeholders/company-1.jpg'},
        {avatarUrl: '/placeholders/company-2.jpg'},
        {avatarUrl: '/placeholders/company-3.jpg'},
      ],
      tags: [
        {
          hash: 'sha265:8f441db4a6dc00a1d5d9fe7eee9e222d17d05695cd6970cd7ea8687a25411982',
          createdAt: '2025-02-25T09:25:21Z',
          downloadsTotal: 879,
          downloadedByCount: 79,
          downloadedByUsers: [
            {avatarUrl: '/placeholders/company-1.jpg'},
            {avatarUrl: '/placeholders/company-2.jpg'},
            {avatarUrl: '/placeholders/company-3.jpg'},
          ],
          labels: [{name: '1.2.1'}],
          vulnerabilities: [],
        },
        {
          hash: 'sha265:bdef5adfc7661eb7719c164a2167d67405e4ce2b3a36c98e64e8755883aeab39',
          createdAt: '2025-02-25T09:25:21Z',
          sbom: 'aaaaaaaaaaaaa',
          downloadsTotal: 468,
          downloadedByCount: 79,
          downloadedByUsers: [
            {avatarUrl: '/placeholders/company-1.jpg'},
            {avatarUrl: '/placeholders/company-2.jpg'},
            {avatarUrl: '/placeholders/company-3.jpg'},
          ],
          labels: [{name: '1.2.0'}],
          vulnerabilities: [
            {
              id: 'CVE-2025-375',
              severity: [{type: 'CVSS_V4', score: 'CVSS:4.0/AV:N/AC:L/AT:N/PR:N/UI:N/VC:L/VI:L/VA:N/SC:N/SI:N/SA:N'}],
            },
          ],
        },
      ],
    },
  ];

  public list(): Observable<Artifact[]> {
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
