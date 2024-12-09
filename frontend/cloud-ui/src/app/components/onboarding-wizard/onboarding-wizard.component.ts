import {Component, ElementRef, EventEmitter, inject, Output, ViewChild} from '@angular/core';
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
import {
  filter,
  find,
  firstValueFrom,
  last,
  lastValueFrom,
  Observable,
  Subject,
  switchMap,
  takeUntil,
  tap,
  withLatestFrom,
} from 'rxjs';
import {DeploymentService} from '../../services/deployment.service';
import {Application} from '../../types/application';
import {DeploymentTarget} from '../../types/deployment-target';
import {ConnectInstructionsComponent} from '../connect-instructions/connect-instructions.component';

@Component({
  selector: 'app-onboarding-wizard',
  templateUrl: './onboarding-wizard.component.html',
  imports: [
    FaIconComponent,
    OnboardingWizardStepperComponent,
    CdkStep,
    ReactiveFormsModule,
    ConnectInstructionsComponent,
  ],
  animations: [modalFlyInOut],
})
export class OnboardingWizardComponent {
  protected readonly xmarkIcon = faXmark;
  private applications = inject(ApplicationsService);
  private deploymentTargets = inject(DeploymentTargetsService);
  private deployments = inject(DeploymentService);
  @ViewChild('stepper') stepper!: CdkStepper;

  createdApp?: Application;
  createdDeploymentTarget?: DeploymentTarget;

  @Output('closed') closed = new EventEmitter<void>();

  applicationForm = new FormGroup({
    type: new FormControl<string>('sample', Validators.required),
    custom: new FormGroup({
      name: new FormControl<string>(
        {
          value: '',
          disabled: true,
        },
        Validators.required
      ),
      versionName: new FormControl<string>(
        {
          value: '',
          disabled: true,
        },
        Validators.required
      ),
    }),
  });
  fileToUpload: File | null = null;
  @ViewChild('fileInput')
  fileInput?: ElementRef;

  deploymentTargetForm = new FormGroup({
    customerName: new FormControl<string>('', Validators.required),
    accessType: new FormControl<string>('', Validators.required),
    technicalContact: new FormGroup({
      name: new FormControl<string>(
        {
          value: '',
          disabled: true,
        },
        Validators.required
      ),
      email: new FormControl<string>({
        value: '',
        disabled: true,
      }),
    }),
  });

  private loading = false;

  ngOnInit() {
    this.applicationForm.controls.type.valueChanges.pipe(takeUntil(this.destroyed$)).subscribe((type) => {
      if (type === 'sample') {
        // disable validators
        this.applicationForm.controls.custom.controls.name.disable();
        this.applicationForm.controls.custom.controls.versionName.disable();
      } else {
        this.applicationForm.controls.custom.controls.name.enable();
        this.applicationForm.controls.custom.controls.versionName.enable();
      }
    });
    this.deploymentTargetForm.controls.accessType.valueChanges.pipe(takeUntil(this.destroyed$)).subscribe((type) => {
      if (type === 'full') {
        // disable validators
        this.deploymentTargetForm.controls.technicalContact.controls.name.disable();
        this.deploymentTargetForm.controls.technicalContact.controls.email.disable();
      } else {
        this.deploymentTargetForm.controls.technicalContact.controls.name.enable();
        this.deploymentTargetForm.controls.technicalContact.controls.email.enable();
      }
    });
  }

  private destroyed$: Subject<void> = new Subject();
  ngOnDestroy() {
    this.destroyed$.complete();
  }

  onFileSelected(event: Event) {
    this.fileToUpload = (event.target as HTMLInputElement).files?.[0] ?? null;
  }

  attemptContinue() {
    if (this.loading) {
      return;
    }

    if (this.stepper.selectedIndex === 0) {
      if (this.applicationForm.valid) {
        if (this.applicationForm.controls.type.value === 'sample') {
          this.loading = true;
          this.applications.createSample().subscribe((app) => {
            this.createdApp = app;
            this.nextStep();
          });
        } else if (this.fileToUpload != null) {
          this.loading = true;
          this.applications
            .create({
              name: this.applicationForm.controls.custom.controls.name.value!,
              type: 'docker',
            })
            .pipe(
              switchMap((application) =>
                this.applications.createApplicationVersion(
                  application,
                  {
                    name: this.applicationForm.controls.custom.controls.versionName.value!,
                  },
                  this.fileToUpload!
                )
              ),
              withLatestFrom(this.applications.list())
            )
            .subscribe(([version, apps]) => {
              this.createdApp = apps.find((a) => a.id === version.applicationId);
              this.nextStep();
            });
        }
      } else {
        this.applicationForm.markAllAsTouched();
      }
    } else if (this.stepper.selectedIndex === 1) {
      if (this.deploymentTargetForm.valid) {
        this.loading = true;
        const base = {
          name: this.deploymentTargetForm.controls.customerName.value!,
          type: 'docker',
          geolocation: {
            lat: 48.1956026,
            lon: 16.3633028,
          },
        };
        this.deploymentTargets
          .create({
            ...base,
            name: base.name + ' (staging)',
          })
          .pipe(
            /*switchMap(() =>
              this.deploymentTargets.create({
                ...base,
                name: base.name + '-prod',
              })
            ),*/
            tap((dt) => (this.createdDeploymentTarget = dt)),
            switchMap((dt) =>
              this.deployments.create({
                applicationVersionId: this.createdApp!.versions![0].id!,
                deploymentTargetId: dt.id!,
              })
            )
          )
          .subscribe(() => this.nextStep());
      } else {
        this.deploymentTargetForm.markAllAsTouched();
      }
    } else if (this.stepper.selectedIndex == 2) {
      this.close();
    }
  }

  close() {
    this.closed.emit();
  }

  private nextStep() {
    this.loading = false;
    this.stepper.next();
  }
}
