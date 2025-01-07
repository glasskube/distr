import {Component, ElementRef, EventEmitter, inject, OnDestroy, OnInit, Output, ViewChild} from '@angular/core';
import {DeploymentTargetsService} from '../../services/deployment-targets.service';
import {faXmark} from '@fortawesome/free-solid-svg-icons';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {Form, FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {modalFlyInOut} from '../../animations/modal';
import {OnboardingWizardStepperComponent} from './onboarding-wizard-stepper.component';
import {CdkStep, CdkStepper} from '@angular/cdk/stepper';
import {ApplicationsService} from '../../services/applications.service';
import {
  combineLatest,
  first,
  firstValueFrom,
  from,
  last,
  Subject,
  switchMap,
  take,
  takeUntil,
  tap,
  withLatestFrom,
} from 'rxjs';
import {DeploymentService} from '../../services/deployment.service';
import {Application, ApplicationVersion} from '../../types/application';
import {DeploymentTarget} from '../../types/deployment-target';
import {ConnectInstructionsComponent} from '../connect-instructions/connect-instructions.component';
import {UsersService} from '../../services/users.service';
import {OnboardingWizardIntroComponent} from './intro/onboarding-wizard-intro.component';
import {disableControlsWithoutEvent, enableControlsWithoutEvent} from '../../../util/forms';
import {DeploymentType, HelmChartType} from '../../types/deployment';

@Component({
  selector: 'app-onboarding-wizard',
  templateUrl: './onboarding-wizard.component.html',
  imports: [
    FaIconComponent,
    OnboardingWizardStepperComponent,
    CdkStep,
    ReactiveFormsModule,
    ConnectInstructionsComponent,
    OnboardingWizardIntroComponent,
  ],
  animations: [modalFlyInOut],
})
export class OnboardingWizardComponent implements OnInit, OnDestroy {
  protected readonly xmarkIcon = faXmark;

  private readonly applications = inject(ApplicationsService);
  private readonly deploymentTargets = inject(DeploymentTargetsService);
  private readonly deployments = inject(DeploymentService);
  private readonly users = inject(UsersService);

  @ViewChild('stepper') stepper!: CdkStepper;

  app?: Application;
  createdDeploymentTarget?: DeploymentTarget;

  @Output('closed') closed = new EventEmitter<void>();

  introForm = new FormGroup({});

  applicationForm = new FormGroup({
    sampleApplication: new FormControl<boolean>(false, {nonNullable: true}),
    type: new FormControl<DeploymentType | null>(null, Validators.required),
    docker: new FormGroup({
      name: new FormControl<string>('', Validators.required),
      versionName: new FormControl<string>('', Validators.required),
    }),
    kubernetes: new FormGroup({
      name: new FormControl<string>('', Validators.required),
      versionName: new FormControl<string>('', Validators.required),
      chartType: new FormControl<HelmChartType>('repository', {
        nonNullable: true,
        validators: Validators.required,
      }),
      chartName: new FormControl<string>('', Validators.required),
      chartUrl: new FormControl<string>('', Validators.required),
      chartVersion: new FormControl<string>('', Validators.required),
    }),
  });
  dockerComposeFile: File | null = null;
  @ViewChild('dockerComposeFileInput')
  dockerComposeFileInput?: ElementRef;

  baseValuesFile: File | null = null;
  @ViewChild('baseValuesFileInput')
  baseValuesFileInput?: ElementRef;

  templateFile: File | null = null;
  @ViewChild('templateFileInput')
  templateFileInput?: ElementRef;

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
    this.toggleTypeSpecificFields(null);

    this.applicationForm.controls.sampleApplication.valueChanges
      .pipe(takeUntil(this.destroyed$))
      .subscribe((selected) => {
        if (selected) {
          this.applicationForm.controls.type.disable();
          this.toggleTypeSpecificFields(null);
        } else {
          this.applicationForm.controls.type.enable();
          this.toggleTypeSpecificFields(this.applicationForm.controls.type.value);
        }
      });

    this.applicationForm.controls.type.valueChanges.pipe(takeUntil(this.destroyed$)).subscribe((type) => {
      this.toggleTypeSpecificFields(type);
    });

    this.applicationForm.controls.kubernetes.controls.chartType.valueChanges
      .pipe(takeUntil(this.destroyed$))
      .subscribe((type) => {
        if (type === 'repository') {
          this.applicationForm.controls.kubernetes.controls.chartName.enable();
        } else {
          this.applicationForm.controls.kubernetes.controls.chartName.disable();
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

  private toggleTypeSpecificFields(type: DeploymentType | null) {
    switch (type) {
      case 'docker':
        enableControlsWithoutEvent(this.applicationForm.controls.docker);
        disableControlsWithoutEvent(this.applicationForm.controls.kubernetes);
        break;
      case 'kubernetes':
        disableControlsWithoutEvent(this.applicationForm.controls.docker);
        enableControlsWithoutEvent(this.applicationForm.controls.kubernetes);
        if (this.applicationForm.controls.kubernetes.controls.chartType.value === 'oci') {
          this.applicationForm.controls.kubernetes.controls.chartName.disable();
        }
        break;
      default:
        disableControlsWithoutEvent(this.applicationForm.controls.docker);
        disableControlsWithoutEvent(this.applicationForm.controls.kubernetes);
    }
  }

  private destroyed$: Subject<void> = new Subject();

  ngOnDestroy() {
    this.destroyed$.complete();
  }

  onDockerComposeFileSelected(event: Event) {
    this.dockerComposeFile = (event.target as HTMLInputElement).files?.[0] ?? null;
  }

  onBaseValuesFileSelected(event: Event) {
    this.baseValuesFile = (event.target as HTMLInputElement).files?.[0] ?? null;
  }

  onTemplateFileSelected(event: Event) {
    this.templateFile = (event.target as HTMLInputElement).files?.[0] ?? null;
  }

  attemptContinue() {
    if (this.loading) {
      return;
    }

    if (this.stepper.selectedIndex === 0) {
      this.loading = true;

      this.applications
        .list()
        .pipe(first())
        .subscribe((apps) => {
          this.nextStep();
          if (apps.length > 0) {
            this.app = apps[0];
            this.nextStep();
          }
        });
    } else if (this.stepper.selectedIndex === 1) {
      console.log(this.applicationForm);
      if (this.applicationForm.valid) {
        const isDocker = this.applicationForm.controls.type.value === 'docker';
        const fileUploadValid = !isDocker || (isDocker && this.dockerComposeFile != null);
        if (this.applicationForm.controls.sampleApplication.value) {
          this.loading = true;
          this.applications.createSample().subscribe((app) => {
            this.app = app;
            this.nextStep();
          });
        } else if (fileUploadValid) {
          this.loading = true;
          let name;
          let version: ApplicationVersion;
          if (isDocker) {
            name = this.applicationForm.controls.docker.controls.name.value!;
            version = {
              name: this.applicationForm.controls.docker.controls.versionName.value!,
            };
          } else {
            name = this.applicationForm.controls.kubernetes.controls.name.value!;
            const versionFormVal = this.applicationForm.controls.kubernetes.value;
            version = {
              name: versionFormVal.versionName!,
              chartType: versionFormVal.chartType!,
              chartName: versionFormVal.chartName ?? undefined,
              chartUrl: versionFormVal.chartUrl!,
              chartVersion: versionFormVal.chartVersion!,
            };
          }
          this.applications
            .create({
              name: name,
              type: this.applicationForm.controls.type.value!,
            })
            .pipe(
              switchMap((application) => {
                if (isDocker) {
                  return this.applications.createApplicationVersionForDocker(
                    application,
                    version,
                    this.dockerComposeFile!
                  );
                } else {
                  return this.applications.createApplicationVersionForKubernetes(
                    application,
                    version,
                    this.baseValuesFile,
                    this.templateFile
                  );
                }
              }),
              withLatestFrom(this.applications.list())
            )
            .subscribe(([version, apps]) => {
              this.app = apps.find((a) => a.id === version.applicationId);
              this.nextStep();
            });
        }
      } else {
        this.applicationForm.markAllAsTouched();
      }
    } else if (this.stepper.selectedIndex === 2) {
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
                  applicationVersionId: this.app!.versions![0].id!,
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
              applicationName: this.app?.name,
            })
            .subscribe(() => this.nextStep());
        }
      } else {
        this.deploymentTargetForm.markAllAsTouched();
      }
    } else if (this.stepper.selectedIndex == 3) {
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
