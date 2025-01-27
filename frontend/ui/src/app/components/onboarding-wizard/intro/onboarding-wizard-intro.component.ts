import {Component} from '@angular/core';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faArrowRightLong, faBoxes, faContactBook, faServer} from '@fortawesome/free-solid-svg-icons';

@Component({
  selector: 'app-onboarding-wizard-intro',
  imports: [FaIconComponent],
  templateUrl: 'onboarding-wizard-intro.component.html',
})
export class OnboardingWizardIntroComponent {
  protected readonly faBoxes = faBoxes;
  protected readonly faArrowRightLong = faArrowRightLong;
  protected readonly faContactBook = faContactBook;
  protected readonly faServer = faServer;
}
