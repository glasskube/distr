import {Component, inject, ViewChild} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {faB, faBox, faBoxesStacked, faDownload, faLightbulb, faPalette} from '@fortawesome/free-solid-svg-icons';
import {CdkStep, CdkStepper} from '@angular/cdk/stepper';
import {TutorialStepperComponent} from '../stepper/tutorial-stepper.component';
import {OrganizationBrandingService} from '../../services/organization-branding.service';

@Component({
  selector: 'app-branding-tutorial',
  imports: [ReactiveFormsModule, CdkStep, TutorialStepperComponent],
  templateUrl: './branding-tutorial.component.html',
})
export class BrandingTutorialComponent {
  protected readonly faBox = faBox;
  protected readonly faDownload = faDownload;
  protected readonly faPalette = faPalette;
  protected readonly faBoxesStacked = faBoxesStacked;
  protected readonly faB = faB;
  protected readonly faLightbulb = faLightbulb;
  @ViewChild('stepper') private stepper!: CdkStepper;
  protected readonly brandingService = inject(OrganizationBrandingService);
  protected readonly brandingFormGroup = new FormGroup({
    branding: new FormGroup({
      title: new FormControl<string>('', {nonNullable: true}),

    }, Validators.required),
    customerInvite: new FormGroup({}, Validators.required),
    customerConfirmed: new FormControl<boolean>(false, {nonNullable: true})
  });

  // TODO on load, check existing tutorial state and also check if branding already exists and fill form accordingly
  // (customer invite probably can't be checked, because even if one exists, they could have been invited by somebody else)

  protected continueFromWelcome() {
    // TODO put tutorial state
    this.brandingService.get();
    this.stepper.next();
  }

  protected backToWelcome() {
    const oldStep = this.stepper.selected!;
    const wasCompleted = oldStep.completed;
    this.stepper.previous(); // why does this set completed to true if its not submitting ????
    if(!wasCompleted) {
      oldStep.completed = false;
    }
  }

  protected complete() {
    this.brandingFormGroup.markAllAsTouched();
    if(this.brandingFormGroup.valid) {
      // TODO put tutorial state
      this.stepper.selected!.completed = true;
    }
  }

}
