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
import {firstValueFrom, lastValueFrom, switchMap, withLatestFrom} from 'rxjs';
import {DeploymentService} from '../../services/deployment.service';
import {Application} from '../../types/application';

@Component({
  selector: 'app-onboarding-wizard',
  templateUrl: './onboarding-wizard.component.html',
  imports: [FaIconComponent, OnboardingWizardStepperComponent, CdkStep, ReactiveFormsModule],
  animations: [modalFlyInOut],
})
export class OnboardingWizardComponent {
  protected readonly xmarkIcon = faXmark;
  private applications = inject(ApplicationsService);
  private deploymentTargets = inject(DeploymentTargetsService);
  private deployments = inject(DeploymentService);
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
    customerName: new FormControl<string>('', Validators.required),
    accessType: new FormControl<string>('full', Validators.required),
    technicalContact: new FormGroup({
      name: new FormControl<string>({
        value: '',
        disabled: true
      }, Validators.required),
      email: new FormControl<string>({
        value: '',
        disabled: true
      })
    })
  });

  private createdApplicationVersionId?: string;

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

    this.deploymentTargetForm.controls.accessType.valueChanges.subscribe(type => {
      if(type === 'full') {
        // disable validators
        // TODO loop over controls
        this.deploymentTargetForm.controls.technicalContact.controls.name.disable();
        this.deploymentTargetForm.controls.technicalContact.controls.email.disable();
      } else {
        this.deploymentTargetForm.controls.technicalContact.controls.name.enable();
        this.deploymentTargetForm.controls.technicalContact.controls.email.enable();
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
        if(this.applicationForm.controls.type.value === 'sample') {
          this.loading = true;
          this.applications.createSample().subscribe((app) => {
            this.createdApplicationVersionId = app.versions![0].id
            this.loading = false;
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
          })).subscribe((av) => {
            this.createdApplicationVersionId = av.id;
            this.loading = false;
            this.stepper.next()
          })
        }
      }
    } else if(this.stepper.selectedIndex == 1) {
      if(this.deploymentTargetForm.valid) {
        this.loading = true
        const base = {
          name: this.deploymentTargetForm.controls.customerName.value!.toLowerCase().replaceAll(' ', '-'),
          type: "docker",
          geolocation: {
            lat: 48.1956026,
            lon: 16.3633028
          }
        }
        this.deploymentTargets.create({
          ...base,
          name: base.name + "-staging",
        }).pipe(switchMap(() => {
          return this.deploymentTargets.create({
            ...base,
            name: base.name + "-prod"
          })
        }), switchMap(dt => {
          return this.deployments.create({
            applicationVersionId: this.createdApplicationVersionId!,
            deploymentTargetId: dt.id!
          })
        })).subscribe((x) => {
          this.loading = false;
          this.stepper.next();
        })
      }
    } else if(this.stepper.selectedIndex == 2) {
      // TODO
      window.location.href = '/dashboard';
    }
  }
}
