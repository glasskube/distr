import {CdkStep, CdkStepper} from '@angular/cdk/stepper';
import {Component, EventEmitter, inject, OnDestroy, OnInit, Output, signal, ViewChild} from '@angular/core';
import {toObservable} from '@angular/core/rxjs-interop';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faShip, faXmark} from '@fortawesome/free-solid-svg-icons';
import {
  Application,
  ApplicationVersion,
  DeploymentRequest,
  DeploymentTarget,
  DeploymentTargetScope,
  DeploymentType,
} from '@glasskube/distr-sdk';
import {combineLatest, firstValueFrom, map, Subject, takeUntil} from 'rxjs';
import {getFormDisplayedError} from '../../../util/errors';
import {modalFlyInOut} from '../../animations/modal';
import {
  mapToDeploymentRequest,
  DeploymentFormComponent,
  DeploymentFormValue,
} from '../../deployment-form/deployment-form.component';
import {AutotrimDirective} from '../../directives/autotrim.directive';
import {ApplicationsService} from '../../services/applications.service';
import {DeploymentTargetsService} from '../../services/deployment-targets.service';
import {FeatureFlagService} from '../../services/feature-flag.service';
import {ToastService} from '../../services/toast.service';
import {ConnectInstructionsComponent} from '../connect-instructions/connect-instructions.component';
import {InstallationWizardStepperComponent} from './installation-wizard-stepper.component';
import {KUBERNETES_RESOURCE_MAX_LENGTH, KUBERNETES_RESOURCE_NAME_REGEX} from '../../../util/validation';

@Component({
  selector: 'app-installation-wizard',
  templateUrl: './installation-wizard.component.html',
  imports: [
    ReactiveFormsModule,
    FaIconComponent,
    InstallationWizardStepperComponent,
    CdkStep,
    ConnectInstructionsComponent,
    AutotrimDirective,
    DeploymentFormComponent,
  ],
  animations: [modalFlyInOut],
})
export class InstallationWizardComponent implements OnInit, OnDestroy {
  protected readonly xmarkIcon = faXmark;
  protected readonly shipIcon = faShip;

  private readonly toast = inject(ToastService);
  private readonly applications = inject(ApplicationsService);
  private readonly deploymentTargets = inject(DeploymentTargetsService);
  protected readonly featureFlags = inject(FeatureFlagService);

  @ViewChild('stepper') private stepper?: CdkStepper;

  @Output('closed') readonly closed = new EventEmitter<void>();

  readonly deploymentTargetForm = new FormGroup({
    type: new FormControl<DeploymentType>('docker', Validators.required),
    name: new FormControl('', Validators.required),
    namespace: new FormControl(
      {value: '', disabled: true},
      {
        nonNullable: true,
        validators: [
          Validators.required,
          Validators.maxLength(KUBERNETES_RESOURCE_MAX_LENGTH),
          Validators.pattern(KUBERNETES_RESOURCE_NAME_REGEX),
        ],
      }
    ),
    clusterScope: new FormControl(
      {value: false, disabled: true},
      {nonNullable: true, validators: [Validators.required]}
    ),
    scope: new FormControl<DeploymentTargetScope>(
      {value: 'namespace', disabled: true},
      {nonNullable: true, validators: [Validators.required]}
    ),
  });

  readonly agentForm = new FormGroup({});

  readonly deployForm = new FormControl<DeploymentFormValue | undefined>(undefined, Validators.required);

  protected selectedApplication?: Application;
  protected availableApplicationVersions = signal<ApplicationVersion[]>([]);
  protected selectedDeploymentTarget = signal<DeploymentTarget | null>(null);

  readonly applications$ = combineLatest([this.applications.list(), toObservable(this.selectedDeploymentTarget)]).pipe(
    map(([apps, dt]) => apps.filter((app) => app.type === dt?.type))
  );

  private loading = false;
  private readonly destroyed$ = new Subject<void>();

  ngOnInit() {
    this.deploymentTargetForm.controls.type.valueChanges.pipe(takeUntil(this.destroyed$)).subscribe((s) => {
      if (s === 'kubernetes') {
        this.deploymentTargetForm.controls.namespace.enable();
        this.deploymentTargetForm.controls.clusterScope.enable();
        this.deploymentTargetForm.controls.scope.enable();
      } else {
        this.deploymentTargetForm.controls.namespace.disable();
        this.deploymentTargetForm.controls.clusterScope.disable();
        this.deploymentTargetForm.controls.scope.disable();
      }
    });

    this.deploymentTargetForm.controls.clusterScope.valueChanges
      .pipe(takeUntil(this.destroyed$))
      .subscribe((value) => this.deploymentTargetForm.controls.scope.setValue(value ? 'cluster' : 'namespace'));
  }

  ngOnDestroy() {
    this.destroyed$.next();
    this.destroyed$.complete();
  }

  async attemptContinue() {
    if (this.loading) {
      return;
    }

    switch (this.stepper?.selectedIndex) {
      case 0:
        await this.continueFromDeploymentType();
        break;
      case 1:
        this.loading = true;
        this.nextStep();
        break;
      case 2:
        await this.continueFromDeployStep();
        break;
    }
  }

  private async continueFromDeploymentType() {
    this.deploymentTargetForm.markAllAsTouched();
    if (!this.deploymentTargetForm.valid || this.loading) {
      return;
    }

    this.loading = true;
    try {
      const created = await firstValueFrom(
        this.deploymentTargets.create({
          name: this.deploymentTargetForm.value.name!,
          type: this.deploymentTargetForm.value.type!,
          namespace: this.deploymentTargetForm.value.namespace,
          scope: this.deploymentTargetForm.value.scope,
          deployments: [],
        })
      );
      this.selectedDeploymentTarget.set(created as DeploymentTarget);
      this.deployForm.setValue({
        deploymentTargetId: created.id,
      });
      this.nextStep();
    } catch (e) {
      const msg = getFormDisplayedError(e);
      if (msg) {
        this.toast.error(msg);
      }
    } finally {
      this.loading = false;
    }
  }

  private async continueFromDeployStep() {
    this.deployForm.markAllAsTouched();
    if (!this.deployForm.valid || this.loading) {
      return;
    }

    try {
      this.loading = true;
      const deployment = mapToDeploymentRequest(this.deployForm.value!);
      await firstValueFrom(this.deploymentTargets.deploy(deployment));
      this.toast.success('Deployment saved successfully');
      this.close();
    } catch (e) {
      const msg = getFormDisplayedError(e);
      if (msg) {
        this.toast.error(msg);
      }
    } finally {
      this.loading = false;
    }
  }

  close() {
    this.closed.emit();
  }

  private nextStep() {
    this.loading = false;
    this.stepper?.next();
  }
}
