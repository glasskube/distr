import {CdkStepper, CdkStepperModule} from '@angular/cdk/stepper';
import {NgTemplateOutlet} from '@angular/common';
import {Component, EventEmitter, Output} from '@angular/core';
import {FormGroup, ReactiveFormsModule} from '@angular/forms';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faDocker, faServicestack} from '@fortawesome/free-brands-svg-icons';
import {faNetworkWired, faServer, faShip} from '@fortawesome/free-solid-svg-icons';

@Component({
  selector: 'app-deployment-wizard-stepper',
  templateUrl: './deployment-wizard-stepper.component.html',
  providers: [{provide: CdkStepper, useExisting: DeploymentWizardStepperComponent}],
  imports: [CdkStepperModule, ReactiveFormsModule, FaIconComponent, NgTemplateOutlet],
})
export class DeploymentWizardStepperComponent extends CdkStepper {
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

  protected readonly faNetworkWired = faNetworkWired;
  protected readonly faServicestack = faServicestack;
  protected readonly faServer = faServer;
}
