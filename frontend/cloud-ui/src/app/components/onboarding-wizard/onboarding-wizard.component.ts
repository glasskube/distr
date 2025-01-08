import {CdkStep, CdkStepper} from '@angular/cdk/stepper';
import {Component, ElementRef, EventEmitter, inject, OnDestroy, OnInit, Output, ViewChild} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faXmark} from '@fortawesome/free-solid-svg-icons';
import {firstValueFrom, Subject, takeUntil} from 'rxjs';
import {disableControlsWithoutEvent, enableControlsWithoutEvent} from '../../../util/forms';
import {modalFlyInOut} from '../../animations/modal';
import {ApplicationsService} from '../../services/applications.service';
import {DeploymentTargetsService} from '../../services/deployment-targets.service';
import {DeploymentService} from '../../services/deployment.service';
import {CreateUserAccountRequest, UsersService} from '../../services/users.service';
import {Application, ApplicationVersion} from '../../types/application';
import {Deployment, DeploymentType, HelmChartType} from '../../types/deployment';
import {DeploymentTarget} from '../../types/deployment-target';
import {ConnectInstructionsComponent} from '../connect-instructions/connect-instructions.component';
import {OnboardingWizardIntroComponent} from './intro/onboarding-wizard-intro.component';
import {OnboardingWizardStepperComponent} from './onboarding-wizard-stepper.component';

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
      chartName: new FormControl<string>('', {nonNullable: true, validators: [Validators.required]}),
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
    namespace: new FormControl<string>(
      {value: '', disabled: true},
      {nonNullable: true, validators: [Validators.required]}
    ),
    releaseName: new FormControl<string>(
      {value: '', disabled: true},
      {nonNullable: true, validators: [Validators.required]}
    ),
    valuesYaml: new FormControl<string>(
      {value: '', disabled: true},
      {nonNullable: true, validators: [Validators.required]}
    ),
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
        if (this.app?.type === 'kubernetes') {
          this.deploymentTargetForm.controls.namespace.enable();
          this.deploymentTargetForm.controls.releaseName.enable();
          this.deploymentTargetForm.controls.valuesYaml.enable();
        }
      } else {
        this.deploymentTargetForm.controls.technicalContactEmail.enable();
        this.deploymentTargetForm.controls.namespace.disable();
        this.deploymentTargetForm.controls.releaseName.disable();
        this.deploymentTargetForm.controls.valuesYaml.disable();
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

  async attemptContinue() {
    if (this.loading) {
      return;
    }

    if (this.stepper.selectedIndex === 0) {
      this.loading = true;
      const apps = await firstValueFrom(this.applications.list());
      this.nextStep();
      if ((apps.length > 0 && apps[0].versions?.length) ?? 0 > 0) {
        this.app = apps[0];
        await this.prepareFormAfterApplicationCreated(this.app, this.app.versions![0]);
        this.nextStep();
      }
    } else if (this.stepper.selectedIndex === 1) {
      if (this.applicationForm.valid) {
        const isDocker = this.applicationForm.controls.type.value === 'docker';
        const fileUploadValid = !isDocker || this.dockerComposeFile !== null;
        if (this.applicationForm.controls.sampleApplication.value) {
          this.loading = true;
          this.app = await firstValueFrom(this.applications.createSample());
          this.nextStep();
        } else if (fileUploadValid) {
          this.loading = true;
          this.app = await firstValueFrom(this.applications.create(this.getApplicationForSubmit()));
          const createdVersion = await firstValueFrom(
            isDocker
              ? this.applications.createApplicationVersionForDocker(
                  this.app,
                  this.getApplicationVersionForSubmit(),
                  this.dockerComposeFile!
                )
              : this.applications.createApplicationVersionForKubernetes(
                  this.app,
                  this.getApplicationVersionForSubmit(),
                  this.baseValuesFile,
                  this.templateFile
                )
          );
          await this.prepareFormAfterApplicationCreated(this.app, createdVersion);
          this.nextStep();
        }
      } else {
        this.applicationForm.markAllAsTouched();
      }
    } else if (this.stepper.selectedIndex === 2) {
      if (this.deploymentTargetForm.valid) {
        this.loading = true;
        if (this.deploymentTargetForm.value.accessType === 'full') {
          this.createdDeploymentTarget = await firstValueFrom(
            this.deploymentTargets.create(this.getDeploymentTargetForSubmit())
          );
          await firstValueFrom(this.deployments.create(this.getDeploymentForSubmit()));
          this.nextStep();
        } else {
          await firstValueFrom(this.users.addUser(this.getUserAccountForSubmit()));
          this.nextStep();
        }
      } else {
        this.deploymentTargetForm.markAllAsTouched();
      }
    } else if (this.stepper.selectedIndex === 3) {
      this.close();
    }
  }

  private async prepareFormAfterApplicationCreated(app: Application, version: ApplicationVersion) {
    if (app.type === 'kubernetes') {
      this.deploymentTargetForm.controls.namespace.enable();
      this.deploymentTargetForm.controls.valuesYaml.enable();
      this.deploymentTargetForm.controls.releaseName.enable();
      const valuesYaml = (await firstValueFrom(this.applications.getTemplateFile(app.id!, version.id!))) ?? undefined;
      const releaseName = app.name!.trim().toLowerCase().replace(/\W+/g, '-');
      this.deploymentTargetForm.patchValue({valuesYaml, releaseName});
    }
  }

  private getApplicationForSubmit(): Application {
    return {
      name:
        this.applicationForm.value.type === 'docker'
          ? this.applicationForm.controls.docker.controls.name.value!
          : this.applicationForm.controls.kubernetes.controls.name.value!,
      type: this.applicationForm.controls.type.value!,
    };
  }

  private getApplicationVersionForSubmit(): ApplicationVersion {
    if (this.app?.type === 'docker') {
      return {
        name: this.applicationForm.controls.docker.controls.versionName.value!,
      };
    } else {
      const versionFormVal = this.applicationForm.controls.kubernetes.value;
      return {
        name: versionFormVal.versionName!,
        chartType: versionFormVal.chartType!,
        chartName: versionFormVal.chartName,
        chartUrl: versionFormVal.chartUrl!,
        chartVersion: versionFormVal.chartVersion!,
      };
    }
  }

  private getDeploymentTargetForSubmit(): DeploymentTarget {
    return {
      name: this.deploymentTargetForm.value.customerName! + ' (staging)',
      type: this.app!.type,
      namespace: this.deploymentTargetForm.value.namespace,
      geolocation: {
        lat: 48.1956026,
        lon: 16.3633028,
      },
    };
  }

  private getDeploymentForSubmit(): Deployment {
    return {
      applicationVersionId: this.app!.versions![0].id!,
      deploymentTargetId: this.createdDeploymentTarget!.id!,
      releaseName: this.deploymentTargetForm.value.releaseName,
      valuesYaml: this.deploymentTargetForm.value.valuesYaml
        ? btoa(this.deploymentTargetForm.value.valuesYaml)
        : undefined,
    };
  }

  private getUserAccountForSubmit(): CreateUserAccountRequest {
    return {
      email: this.deploymentTargetForm.value.technicalContactEmail!,
      name: this.deploymentTargetForm.value.customerName!,
      userRole: 'customer',
      applicationName: this.app?.name,
    };
  }

  close() {
    this.closed.emit();
  }

  private nextStep() {
    this.loading = false;
    this.stepper.next();
  }
}
