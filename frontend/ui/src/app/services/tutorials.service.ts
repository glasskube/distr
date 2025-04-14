import {inject, Injectable} from '@angular/core';
import {HttpClient} from '@angular/common/http';
import {Tutorial, TutorialProgress, TutorialProgressRequest} from '../types/tutorials';
import {map, Observable, shareReplay, Subject, startWith, switchMap} from 'rxjs';
import {IconDefinition} from '@fortawesome/angular-fontawesome';
import {faBox, faBoxesStacked, faPalette} from '@fortawesome/free-solid-svg-icons';

interface TutorialView {
  id: Tutorial;
  name: string;
  icon: IconDefinition;
  progress?: TutorialProgress;
}

@Injectable({providedIn: 'root'})
export class TutorialsService {
  protected readonly faBox = faBox;
  protected readonly faPalette = faPalette;
  protected readonly faBoxesStacked = faBoxesStacked;
  private readonly baseUrl = '/api/v1/tutorial-progress';
  private readonly httpClient = inject(HttpClient);

  protected readonly tutorials: TutorialView[] = [
    {
      name: 'Branding and Customer Portal',
      id: 'branding',
      icon: this.faPalette,
    },
    {
      name: 'Applications and Agents',
      id: 'agents',
      icon: this.faBoxesStacked,
    },
    {
      name: 'Artifact Registry',
      id: 'registry',
      icon: this.faBox,
    },
  ];

  private readonly refresh$ = new Subject<void>();
  public readonly tutorialsProgress$ = this.refresh$.pipe(
    startWith(undefined),
    switchMap(() => this.list()),
    map((progresses) => {
      return this.tutorials.map((t) => {
        const progress = progresses.find((p) => p.tutorial === t.id);
        if (progress) {
          return {
            ...t,
            progress,
          };
        } else {
          return t;
        }
      });
    }),
    shareReplay(1)
  );

  public readonly notAllStarted$ = this.tutorialsProgress$.pipe(
    map((tutorials) => tutorials.some((t) => !t.progress?.createdAt))
  );

  public readonly allCompleted$ = this.tutorialsProgress$.pipe(
    map((tutorials) => !tutorials.some((t) => !t.progress?.completedAt))
  );

  private list(): Observable<TutorialProgress[]> {
    return this.httpClient.get<TutorialProgress[]>(`${this.baseUrl}`);
  }

  public refreshList() {
    this.refresh$.next();
  }

  public get(tutorial: Tutorial): Observable<TutorialProgress> {
    return this.httpClient.get<TutorialProgress>(`${this.baseUrl}/${tutorial}`);
  }

  public save(tutorial: Tutorial, progress: TutorialProgressRequest): Observable<TutorialProgress> {
    return this.httpClient.put<TutorialProgress>(`${this.baseUrl}/${tutorial}`, progress);
  }
}
