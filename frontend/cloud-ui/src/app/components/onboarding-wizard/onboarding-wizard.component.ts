import {Component, ElementRef, EventEmitter, inject, Output, ViewChild} from '@angular/core';
import {DeploymentTargetsService} from '../../services/deployment-targets.service';
import {faXmark} from '@fortawesome/free-solid-svg-icons';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {modalFlyInOut} from '../../animations/modal';
import {OnboardingWizardStepperComponent} from './onboarding-wizard-stepper.component';
import {CdkStep, CdkStepper} from '@angular/cdk/stepper';
import {ApplicationsService} from '../../services/applications.service';
import {Subject, switchMap, takeUntil, tap, withLatestFrom} from 'rxjs';
import {DeploymentService} from '../../services/deployment.service';
import {Application} from '../../types/application';
import {DeploymentTarget} from '../../types/deployment-target';
import {ConnectInstructionsComponent} from '../connect-instructions/connect-instructions.component';
import {UsersService} from '../../services/users.service';

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

  private readonly applications = inject(ApplicationsService);
  private readonly deploymentTargets = inject(DeploymentTargetsService);
  private readonly deployments = inject(DeploymentService);
  private readonly users = inject(UsersService);

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
    accessType: new FormControl<'full' | 'none'>('full', Validators.required),
    technicalContactEmail: new FormControl<string>({value: '', disabled: true}, [
      Validators.required,
      Validators.email,
    ]),
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
        this.deploymentTargetForm.controls.technicalContactEmail.disable();
      } else {
        this.deploymentTargetForm.controls.technicalContactEmail.enable();
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
        if (this.deploymentTargetForm.value.accessType === 'full') {
          this.deploymentTargets
            .create({
              name: this.deploymentTargetForm.controls.customerName.value! + ' (staging)',
              type: 'docker',
              geolocation: {
                lat: 48.1956026,
                lon: 16.3633028,
              },
            })
            .pipe(
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
          this.users
            .addUser({
              email: this.deploymentTargetForm.value.technicalContactEmail!,
              name: this.deploymentTargetForm.value.customerName!,
              userRole: 'customer',
              applicationName: this.createdApp?.name
            })
            .subscribe(() => this.nextStep());
        }
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
