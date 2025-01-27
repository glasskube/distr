import {Component, EventEmitter, Output} from '@angular/core';
import {CdkStepper, CdkStepperModule} from '@angular/cdk/stepper';
import {NgTemplateOutlet} from '@angular/common';
import {FormGroup, ReactiveFormsModule} from '@angular/forms';
import {faAddressBook, faCube} from '@fortawesome/free-solid-svg-icons';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faDocker, faHubspot} from '@fortawesome/free-brands-svg-icons';

@Component({
  selector: 'app-onboarding-wizard-stepper',
  templateUrl: './onboarding-wizard-stepper.component.html',
  providers: [{provide: CdkStepper, useExisting: OnboardingWizardStepperComponent}],
  imports: [NgTemplateOutlet, CdkStepperModule, ReactiveFormsModule, FaIconComponent],
})
export class OnboardingWizardStepperComponent extends CdkStepper {
  protected readonly faDocker = faDocker;
  protected readonly faAddressBook = faAddressBook;
  protected readonly faHubspot = faHubspot;
  protected readonly faCube = faCube;

  @Output('attemptContinue') attemptContinueOutput: EventEmitter<void> = new EventEmitter();

  currentFormGroup() {
    return this.selected!.stepControl as FormGroup;
  }
}
