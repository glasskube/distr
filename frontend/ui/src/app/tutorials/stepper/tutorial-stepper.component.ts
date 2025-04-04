import {Component} from '@angular/core';
import {CdkStepper, CdkStepperModule} from '@angular/cdk/stepper';
import {ReactiveFormsModule} from '@angular/forms';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {NgTemplateOutlet} from '@angular/common';
import {faCircle, faCircleCheck} from '@fortawesome/free-regular-svg-icons';

/*
class TutorialStep extends CdkStep {
  readonly id: string;

  constructor(id: string) {
    super();
    this.id = id;
  }
}

class TutorialCdkStepper extends CdkStepper {
  override readonly steps: QueryList<TutorialStep>;

  constructor(steps: QueryList<TutorialStep>) {
    super();
    this.steps = steps;
  }
}
*/

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
