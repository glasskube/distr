import {Component, inject} from '@angular/core';
import {ReactiveFormsModule} from '@angular/forms';
import {RouterLink} from '@angular/router';
import {FaIconComponent, IconDefinition} from '@fortawesome/angular-fontawesome';
import {
  faArrowRight,
  faB,
  faBox,
  faBoxesStacked,
  faCheck,
  faDownload,
  faLightbulb,
  faPalette,
} from '@fortawesome/free-solid-svg-icons';
import {Tutorial, TutorialProgress} from '../types/tutorials';
import {
  catchError,
  concatMap,
  from,
  map,
  Observable,
  of,
  reduce,
  scan,
  shareReplay,
  startWith,
  switchMap,
  tap,
  toArray,
} from 'rxjs';
import {TutorialsService} from '../services/tutorials.service';
import {AsyncPipe} from '@angular/common';
import {UuidComponent} from '../components/uuid';

interface TutorialView {
  id: Tutorial;
  name: string;
  icon: IconDefinition;
  progress?: TutorialProgress;
}

@Component({
  selector: 'app-tutorials',
  imports: [ReactiveFormsModule, FaIconComponent, RouterLink, AsyncPipe],
  templateUrl: './tutorials.component.html',
})
export class TutorialsComponent {
  protected readonly faBox = faBox;
  protected readonly faPalette = faPalette;
  protected readonly faBoxesStacked = faBoxesStacked;
  protected readonly faLightbulb = faLightbulb;
  protected readonly faArrowRight = faArrowRight;
  protected readonly faCheck = faCheck;
  private readonly tutorialsService = inject(TutorialsService);

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

  protected readonly tutorialsWithProgress$ = from(this.tutorials).pipe(
    concatMap((t) => {
      return this.tutorialsService.get(t.id).pipe(
        map((progress) => {
          return {
            ...t,
            progress,
          } as TutorialView;
        }),
        catchError(() => of(t))
      );
    }),
    reduce((acc, val) => [...acc, val], [] as TutorialView[]),
    shareReplay(1)
  );

  protected readonly allCompleted$ = this.tutorialsWithProgress$.pipe(
    map((tutorials) => !tutorials.some((t) => !t.progress?.completedAt)),
    tap(console.log)
  );
}
