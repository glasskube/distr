import {Component, EventEmitter, Output} from '@angular/core';
import {CdkStep, CdkStepper, CdkStepperModule} from '@angular/cdk/stepper';
import {NgTemplateOutlet} from '@angular/common';
import {Form, FormGroup, ReactiveFormsModule} from '@angular/forms';
import {faNetworkWired, faShip, faXmark} from '@fortawesome/free-solid-svg-icons';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faDocker} from '@fortawesome/free-brands-svg-icons';

@Component({
  selector: 'app-onboarding-wizard-stepper',
  templateUrl: './onboarding-wizard-stepper.component.html',
  providers: [{provide: CdkStepper, useExisting: OnboardingWizardStepperComponent}],
  imports: [NgTemplateOutlet, CdkStepperModule, ReactiveFormsModule, FaIconComponent],
})
export class OnboardingWizardStepperComponent extends CdkStepper {
  protected readonly dockerIcon = faDocker;
  protected readonly shipIcon = faShip;
  protected readonly networkIcon = faNetworkWired;
  @Output('attemptContinue') attemptContinueOutput: EventEmitter<void> = new EventEmitter();

  currentFormGroup() {
    return this.selected!.stepControl as FormGroup;
  }

  isApplicationStep(): boolean {
    return this.selectedIndex === 0;
  }

  isDeploymentTargetStep(): boolean {
    return this.selectedIndex === 1;
  }

  isFinalStep(): boolean {
    return this.selectedIndex === this.steps.length - 1;
  }
}
