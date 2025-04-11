import {Component} from '@angular/core';
import {CdkStepper, CdkStepperModule} from '@angular/cdk/stepper';
import {ReactiveFormsModule} from '@angular/forms';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {NgTemplateOutlet} from '@angular/common';
import {faCircle, faCircleCheck} from '@fortawesome/free-regular-svg-icons';

@Component({
  selector: 'app-tutorial-stepper',
  templateUrl: './tutorial-stepper.component.html',
  providers: [{provide: CdkStepper, useExisting: TutorialStepperComponent}],
  imports: [CdkStepperModule, ReactiveFormsModule, FaIconComponent, NgTemplateOutlet],
})
export class TutorialStepperComponent extends CdkStepper {
  protected readonly faCircle = faCircle;
  protected readonly faCircleCheck = faCircleCheck;

  protected isCurrentStep(i: number): boolean {
    return this.selectedIndex === i;
  }
}
