import {inject, Injectable} from '@angular/core';
import {from, Observable} from 'rxjs';
import {digestMessage} from '../../util/crypto';
import {AuthService} from './auth.service';

export interface HasDownloads {
  downloadsTotal: number;
  downloadedByCount: number;
  downloadedByUsers: ArtifactUser[];
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
  lastScannedAt?: string;
}

export interface ArtifactWithTags extends Artifact {
  tags: ArtifactTag[];
}

@Injectable({providedIn: 'root'})
export class ArtifactsService {
  private readonly auth = inject(AuthService);

  private async getArtifacts(): Promise<ArtifactWithTags[]> {
    return [
      {
        id: '86d0f4c2-650c-480f-b875-ae2857c9753f',
        name: 'distr',
        downloadsTotal: 40,
        downloadedByCount: this.auth.hasRole('vendor') ? 13 : 1,
        downloadedByUsers: await this.getDownloadedByUsers(),
        tags: [
          {
            id: 'b63e6df0-0e78-4c93-8543-db0926967411',
            hash: 'sha265:78f8664cbfbec1c378f8c2af68f6fcbb1ce3faf1388c9d0b70533152b1415e98',
            sbom: 'aaaaaaaaaaaaa',
            createdAt: '2025-02-25T09:25:21Z',
            downloadsTotal: 16,
            downloadedByCount: this.auth.hasRole('vendor') ? 12 : 1,
            downloadedByUsers: await this.getDownloadedByUsers(),
            labels: [{name: 'latest'}, {name: '1.2.1'}],
            vulnerabilities: [],
            lastScannedAt: '2025-02-25T09:25:21Z',
          },
          {
            id: 'cdf206ae-91c4-43f6-b116-7a28e083d9c8',
            hash: 'sha265:28b7a85914586d15a531566443b6d5ea6d11ad38b1e75fa753385f03b0a0a57f',
            createdAt: '2025-02-25T09:25:21Z',
            downloadsTotal: 24,
            downloadedByCount: this.auth.hasRole('vendor') ? 1 : 1,
            downloadedByUsers: await this.getDownloadedByUsers(true, 1),
            labels: [{name: '1.1.6'}],
            sbom: 'aaaaaaaaaaaaa',
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
            lastScannedAt: '2025-02-25T09:25:21Z',
          },
        ],
      },
      {
        id: '6c988429-cef7-45ce-9ef5-a4af55cfc8a2',
        name: 'distr/docker-agent',
        downloadsTotal: 1234,
        downloadedByCount: this.auth.hasRole('vendor') ? 759 : 1,
        downloadedByUsers: await this.getDownloadedByUsers(),
        tags: [
          {
            id: '357d4c97-aead-4b94-b329-fc0670c5ce4c',
            hash: 'sha265:8f441db4a6dc00a1d5d9fe7eee9e222d17d05695cd6970cd7ea8687a25411982',
            createdAt: '2025-02-25T09:25:21Z',
            downloadsTotal: 879,
            downloadedByCount: this.auth.hasRole('vendor') ? 79 : 0,
            downloadedByUsers: await this.getDownloadedByUsers(false),
            labels: [{name: '1.2.1'}],
            vulnerabilities: [],
          },
          {
            id: 'b66a042e-076f-477c-9d7c-9a356f5b34db',
            hash: 'sha265:bdef5adfc7661eb7719c164a2167d67405e4ce2b3a36c98e64e8755883aeab39',
            createdAt: '2025-02-25T09:25:21Z',
            sbom: 'aaaaaaaaaaaaa',
            downloadsTotal: 468,
            downloadedByCount: this.auth.hasRole('vendor') ? 79 : 1,
            downloadedByUsers: await this.getDownloadedByUsers(true),
            labels: [{name: '1.2.0'}],
            vulnerabilities: [
              {
                id: 'CVE-2025-375',
                severity: [{type: 'CVSS_V4', score: 'CVSS:4.0/AV:N/AC:L/AT:N/PR:N/UI:N/VC:L/VI:L/VA:N/SC:N/SI:N/SA:N'}],
              },
            ],
            lastScannedAt: '2025-02-25T09:25:21Z',
          },
        ],
      },
    ];
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
    return from(this.getArtifacts());
  }

  public get(id: string): Observable<ArtifactWithTags> {
    return from(
      (async () => {
        const artifact = (await this.getArtifacts()).find((it) => it.id === id);
        if (artifact !== undefined) {
          return artifact;
        }
        throw new Error('not found');
      })()
    );
  }
}
