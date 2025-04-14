import {Component, inject} from '@angular/core';
import {ReactiveFormsModule} from '@angular/forms';
import {RouterLink} from '@angular/router';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faArrowRight, faCheck, faLightbulb} from '@fortawesome/free-solid-svg-icons';
import {TutorialsService} from '../services/tutorials.service';
import {AsyncPipe} from '@angular/common';

@Component({
  selector: 'app-tutorials',
  imports: [ReactiveFormsModule, FaIconComponent, RouterLink, AsyncPipe],
  templateUrl: './tutorials.component.html',
})
export class TutorialsComponent {
  protected readonly faLightbulb = faLightbulb;
  protected readonly faArrowRight = faArrowRight;
  protected readonly faCheck = faCheck;
  protected readonly tutorialsService = inject(TutorialsService);
}
