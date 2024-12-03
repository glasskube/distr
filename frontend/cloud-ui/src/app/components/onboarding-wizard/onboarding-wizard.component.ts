import {Component, ElementRef, inject, ViewChild} from '@angular/core';
import {GlobeComponent} from '../globe/globe.component';
import {DeploymentTargetsService} from '../../services/deployment-targets.service';
import {ApplicationsComponent} from '../../applications/applications.component';
import {DeploymentTargetsComponent} from '../../deployment-targets/deployment-targets.component';
import {AsyncPipe} from '@angular/common';
import {faShip, faXmark} from '@fortawesome/free-solid-svg-icons';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {FormControl, FormGroup, FormsModule, ReactiveFormsModule, Validators} from '@angular/forms';
import {modalFlyInOut} from '../../animations/modal';
import {OnboardingWizardStepperComponent} from './onboarding-wizard-stepper.component';
import {CdkStep, CdkStepper} from '@angular/cdk/stepper';
import {faDocker} from '@fortawesome/free-brands-svg-icons';
import {ApplicationsService} from '../../services/applications.service';
import {switchMap} from 'rxjs';

@Component({
  selector: 'app-onboarding-wizard',
  templateUrl: './onboarding-wizard.component.html',
  imports: [FaIconComponent, OnboardingWizardStepperComponent, CdkStep, ReactiveFormsModule],
  animations: [modalFlyInOut],
})
export class OnboardingWizardComponent {
  protected readonly xmarkIcon = faXmark;
  private applications = inject(ApplicationsService);
  @ViewChild('stepper') stepper!: CdkStepper;

  applicationForm = new FormGroup({
    type: new FormControl<string>('sample', Validators.required),
    custom: new FormGroup({
      name: new FormControl<string>({
        value: '',
        disabled: true,
      }, Validators.required),
      versionName: new FormControl<string>({
        value: '',
        disabled: true,
      }, Validators.required),
      // TODO file upload
    })
  });
  fileToUpload: File | null = null;
  @ViewChild('fileInput')
  fileInput?: ElementRef;

  deploymentTargetForm = new FormGroup({
    name: new FormControl<string>('', Validators.required),
  });

  installationForm = new FormGroup({

  })

  private applicationDone = false
  private loading = false

  ngOnInit() {
    // TODO subscription kill switch
    this.applicationForm.controls.type.valueChanges.subscribe(type => {
      if(type === 'sample') {
        // disable validators
        // TODO loop over controls
        this.applicationForm.controls.custom.controls.name.disable();
        this.applicationForm.controls.custom.controls.versionName.disable();
      } else {
        this.applicationForm.controls.custom.controls.name.enable();
        this.applicationForm.controls.custom.controls.versionName.enable();
      }
    })
  }

  onFileSelected(event: any) {
    this.fileToUpload = event.target.files[0];
  }

  attemptContinue() {
    if(this.loading) {
      return
    }

    if(this.stepper.selectedIndex == 0) {
      if(this.applicationForm.valid) {
        if(this.applicationDone) {
          this.stepper.next()
        } else if(this.applicationForm.controls.type.value === 'sample') {
          this.applications.createSample().subscribe(() => {
            this.loading = false;
            this.applicationDone = true;
            this.stepper.next()
          });
        } else if(this.fileToUpload != null) {
          this.loading = true
          this.applications.create({
            name: this.applicationForm.controls.custom.controls.name.value!,
            type: "docker"
          }).pipe(switchMap(application => {
            return this.applications.createApplicationVersion(
              application, {
                name: this.applicationForm.controls.custom.controls.versionName.value!,
              }, this.fileToUpload!)
          })).subscribe(() => {
            this.loading = false;
            this.applicationDone = true;
            this.stepper.next()
          })
        }
      }
    } else if(this.stepper.selectedIndex == 1) {
      console.log('TODO handle deployment target')
      this.stepper.next()
    } else if(this.stepper.selectedIndex == 2) {
      // TODO
      window.location.href = '/dashboard';
    }
  }
}
