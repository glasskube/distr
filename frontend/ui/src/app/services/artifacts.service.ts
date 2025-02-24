import {Injectable} from '@angular/core';
import {Observable, of, throwError} from 'rxjs';

export interface Artifact {
  id: string;
  name: string;
  downloads: number;
}

export interface ArtifactTag {
  hash: string;
  downloads: number;
  labels: {name: string}[];
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
      downloads: 782,
      tags: [
        {
          hash: 'sha265:78f8664cbfbec1c378f8c2af68f6fcbb1ce3faf1388c9d0b70533152b1415e98',
          downloads: 345,
          labels: [{name: 'latest'}, {name: '1.2.1'}],
        },
        {
          hash: 'sha265:28b7a85914586d15a531566443b6d5ea6d11ad38b1e75fa753385f03b0a0a57f',
          downloads: 124,
          labels: [{name: '1.1.6'}],
        },
      ],
    },
    {
      id: '6c988429-cef7-45ce-9ef5-a4af55cfc8a2',
      name: 'distr/docker-agent',
      downloads: 1234,
      tags: [
        {
          hash: 'sha265:8f441db4a6dc00a1d5d9fe7eee9e222d17d05695cd6970cd7ea8687a25411982',
          downloads: 879,
          labels: [{name: '1.2.1'}],
        },
        {
          hash: 'sha265:bdef5adfc7661eb7719c164a2167d67405e4ce2b3a36c98e64e8755883aeab39',
          downloads: 468,
          labels: [{name: '1.2.0'}],
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
