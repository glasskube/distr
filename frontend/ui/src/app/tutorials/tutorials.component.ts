import {Component} from '@angular/core';
import {ReactiveFormsModule} from '@angular/forms';
import {RouterLink} from '@angular/router';
import {FaIconComponent, IconDefinition} from '@fortawesome/angular-fontawesome';
import {
  faArrowRight,
  faB,
  faBox,
  faBoxesStacked,
  faDownload,
  faLightbulb,
  faPalette
} from '@fortawesome/free-solid-svg-icons';

interface Tutorial {
  id: string;
  name: string;
  icon: IconDefinition;
}

@Component({
  selector: 'app-tutorials',
  imports: [ReactiveFormsModule, FaIconComponent, RouterLink],
  templateUrl: './tutorials.component.html',
})
export class TutorialsComponent {
  protected readonly faBox = faBox;
  protected readonly faPalette = faPalette;
  protected readonly faBoxesStacked = faBoxesStacked;
  protected readonly faLightbulb = faLightbulb;
  protected readonly faArrowRight = faArrowRight;

  protected readonly tutorials: Tutorial[] = [{
    name: 'Branding and Customer Portal',
    id: 'branding',
    icon: this.faPalette
  }, {
    name: 'Applications and Agents',
    id: 'agents',
    icon: this.faBoxesStacked
  }, {
    name: 'Artifact Registry',
    id: 'registry',
    icon: this.faBox
  }]
}
