import {CdkStepper, CdkStepperModule} from '@angular/cdk/stepper';
import {NgTemplateOutlet} from '@angular/common';
import {Component, EventEmitter, Input, Output} from '@angular/core';
import {FormGroup, ReactiveFormsModule} from '@angular/forms';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faDocker, faServicestack} from '@fortawesome/free-brands-svg-icons';
import {
  faBuildingUser,
  faCog,
  faDharmachakra,
  faNetworkWired,
  faServer,
  faShip,
} from '@fortawesome/free-solid-svg-icons';

@Component({
  selector: 'app-deployment-wizard-stepper',
  templateUrl: './deployment-wizard-stepper.component.html',
  providers: [{provide: CdkStepper, useExisting: DeploymentWizardStepperComponent}],
  imports: [CdkStepperModule, ReactiveFormsModule, FaIconComponent, NgTemplateOutlet],
})
export class DeploymentWizardStepperComponent extends CdkStepper {
  protected readonly dockerIcon = faDocker;
  protected readonly kubernetesIcon = faDharmachakra;
  protected readonly shipIcon = faShip;
  protected readonly networkIcon = faNetworkWired;
  protected readonly buildingUserIcon = faBuildingUser;
  protected readonly cogIcon = faCog;

  @Input() showCustomerStep = false;
  @Output('attemptContinue') attemptContinueOutput: EventEmitter<void> = new EventEmitter();

  currentFormGroup() {
    return this.selected!.stepControl as FormGroup;
  }

  // Adjusted for conditional customer step
  getAdjustedIndex(): number {
    return this.showCustomerStep ? this.selectedIndex : this.selectedIndex + 1;
  }

  isCustomerStep(): boolean {
    return this.showCustomerStep && this.selectedIndex === 0;
  }

  isApplicationStep(): boolean {
    const adjustedIndex = this.getAdjustedIndex();
    return adjustedIndex === 1;
  }

  isTargetStep(): boolean {
    const adjustedIndex = this.getAdjustedIndex();
    return adjustedIndex === 2;
  }

  isConfigurationStep(): boolean {
    const adjustedIndex = this.getAdjustedIndex();
    return adjustedIndex === 3;
  }

  isConnectStep(): boolean {
    const adjustedIndex = this.getAdjustedIndex();
    return adjustedIndex === 4;
  }

  isFinalStep(): boolean {
    return this.selectedIndex === this.steps.length - 1;
  }

  protected readonly faNetworkWired = faNetworkWired;
  protected readonly faServicestack = faServicestack;
  protected readonly faServer = faServer;
}
