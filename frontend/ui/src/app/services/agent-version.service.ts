import {HttpClient} from '@angular/common/http';
import {inject, Injectable} from '@angular/core';
import {Observable, shareReplay} from 'rxjs';
import {AgentVersion} from '@glasskube/distr-sdk';

const baseUrl = '/api/v1/agent-versions';

@Injectable({providedIn: 'root'})
export class AgentVersionService {
  private readonly httpClient = inject(HttpClient);
  private readonly agentVersions$ = this.httpClient.get<AgentVersion[]>(baseUrl).pipe(shareReplay(1));

  public list(): Observable<AgentVersion[]> {
    return this.agentVersions$;
  }
}
