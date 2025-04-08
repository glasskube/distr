import {inject, Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {Tutorial, TutorialProgress, TutorialProgressRequest} from '../types/tutorials';
import {Observable} from 'rxjs';

@Injectable({providedIn: 'root'})
export class TutorialsService {
  private readonly baseUrl = '/api/v1/tutorial-progress';
  private readonly httpClient = inject(HttpClient);

  public get(tutorial: Tutorial): Observable<TutorialProgress> {
    return this.httpClient.get<TutorialProgress>(`${this.baseUrl}/${tutorial}`);
  }

  public save(tutorial: Tutorial, progress: TutorialProgressRequest): Observable<TutorialProgress> {
    return this.httpClient.put<TutorialProgress>(`${this.baseUrl}/${tutorial}`, progress);
  }
}
