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
import {catchError, Observable, of} from 'rxjs';
import {TutorialsService} from '../services/tutorials.service';
import {AsyncPipe} from '@angular/common';
import {UuidComponent} from '../components/uuid';

interface TutorialView {
  id: string;
  name: string;
  icon: IconDefinition;
  progress: Observable<TutorialProgress | undefined>;
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
  private readonly tutorialsService = inject(TutorialsService);

  protected readonly tutorials: TutorialView[] = [
    {
      name: 'Branding and Customer Portal',
      id: 'branding',
      icon: this.faPalette,
      progress: this.tutorialsService.get('branding').pipe(catchError(() => of({tutorial: 'branding' as Tutorial}))),
    },
    {
      name: 'Applications and Agents',
      id: 'agents',
      icon: this.faBoxesStacked,
      progress: this.tutorialsService.get('agents').pipe(catchError(() => of({tutorial: 'agents' as Tutorial}))),
    },
    {
      name: 'Artifact Registry',
      id: 'registry',
      icon: this.faBox,
      progress: this.tutorialsService.get('registry').pipe(catchError(() => of({tutorial: 'registry' as Tutorial}))),
    },
  ];
  protected readonly faCheck = faCheck;
}
