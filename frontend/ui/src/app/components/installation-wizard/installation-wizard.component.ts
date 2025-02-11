import {CdkStep, CdkStepper} from '@angular/cdk/stepper';
import {AsyncPipe} from '@angular/common';
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
import {catchError, combineLatest, firstValueFrom, map, NEVER, shareReplay, Subject, switchMap, takeUntil} from 'rxjs';
import {getFormDisplayedError} from '../../../util/errors';
import {modalFlyInOut} from '../../animations/modal';
import {AutotrimDirective} from '../../directives/autotrim.directive';
import {ApplicationsService} from '../../services/applications.service';
import {AuthService} from '../../services/auth.service';
import {DeploymentTargetsService} from '../../services/deployment-targets.service';
import {FeatureFlagService} from '../../services/feature-flag.service';
import {LicenseService} from '../../services/license.service';
import {ToastService} from '../../services/toast.service';
import {ConnectInstructionsComponent} from '../connect-instructions/connect-instructions.component';
import {YamlEditorComponent} from '../yaml-editor.component';
import {InstallationWizardStepperComponent} from './installation-wizard-stepper.component';

@Component({
  selector: 'app-installation-wizard',
  templateUrl: './installation-wizard.component.html',
  imports: [
    ReactiveFormsModule,
    FaIconComponent,
    InstallationWizardStepperComponent,
    CdkStep,
    AsyncPipe,
    ConnectInstructionsComponent,
    YamlEditorComponent,
    AutotrimDirective,
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
  private readonly licenses = inject(LicenseService);
  private readonly auth = inject(AuthService);

  @ViewChild('stepper') private stepper?: CdkStepper;

  @Output('closed') readonly closed = new EventEmitter<void>();

  readonly deploymentTargetForm = new FormGroup({
    type: new FormControl<DeploymentType>('docker', Validators.required),
    name: new FormControl('', Validators.required),
    namespace: new FormControl({value: '', disabled: true}, {nonNullable: true, validators: [Validators.required]}),
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

  readonly deployForm = new FormGroup({
    deploymentTargetId: new FormControl<string | undefined>(undefined, Validators.required),
    applicationId: new FormControl<string | undefined>(undefined, Validators.required),
    applicationVersionId: new FormControl<string | undefined>({value: undefined, disabled: true}, Validators.required),
    applicationLicenseId: new FormControl<string | undefined>({value: undefined, disabled: true}, Validators.required),
    valuesYaml: new FormControl<string>({value: '', disabled: true}),
    releaseName: new FormControl<string>({value: '', disabled: true}, Validators.required),
    envFileData: new FormControl<string>({value: '', disabled: true}),
  });

  protected selectedApplication?: Application;
  protected availableApplicationVersions: ApplicationVersion[] = [];
  protected selectedDeploymentTarget = signal<DeploymentTarget | null>(null);

  readonly applications$ = combineLatest([this.applications.list(), toObservable(this.selectedDeploymentTarget)]).pipe(
    map(([apps, dt]) => apps.filter((app) => app.type === dt?.type))
  );

  readonly licenses$ = combineLatest([
    this.deployForm.controls.applicationId.valueChanges,
    this.featureFlags.isLicensingEnabled$,
  ]).pipe(
    switchMap(([applicationId, isLicensingEnabled]) =>
      isLicensingEnabled && this.auth.hasRole('customer') && applicationId
        ? this.licenses.getLicensesForApplication(applicationId)
        : NEVER
    ),
    shareReplay(1)
  );

  readonly selectedLicense$ = combineLatest([
    this.deployForm.controls.applicationLicenseId.valueChanges,
    this.licenses$,
  ]).pipe(map(([licenseId, licenses]) => licenses.find((it) => it.id === licenseId)));

  private loading = false;
  private readonly destroyed$ = new Subject<void>();

  ngOnInit() {
    this.featureFlags.isLicensingEnabled$.pipe(takeUntil(this.destroyed$)).subscribe((isLicensingEnabled) => {
      if (isLicensingEnabled && this.auth.hasRole('customer')) {
        this.deployForm.controls.applicationLicenseId.enable();
      }
    });

    combineLatest([
      this.deployForm.controls.applicationId.valueChanges,
      this.applications$,
      this.featureFlags.isLicensingEnabled$,
    ])
      .pipe(takeUntil(this.destroyed$))
      .subscribe(([applicationId, applications, isLicensingEnabled]) => {
        this.updatedSelectedApplication(applications, applicationId);
        if (isLicensingEnabled && this.auth.hasRole('customer')) {
          this.deployForm.controls.applicationVersionId.reset();
        } else {
          this.updateAvailableApplicationVersions(this.selectedApplication?.versions);
        }
      });

    this.deployForm.controls.applicationVersionId.valueChanges
      .pipe(
        takeUntil(this.destroyed$),
        switchMap((id) =>
          id
            ? this.applications.getTemplateFile(this.selectedApplication?.id!, id).pipe(
                catchError(() => NEVER),
                map((data) => [this.selectedApplication?.type, data])
              )
            : NEVER
        )
      )
      .subscribe(([type, templateFile]) => {
        if (type === 'kubernetes') {
          this.deployForm.controls.releaseName.enable();
          this.deployForm.controls.valuesYaml.enable();
          this.deployForm.controls.envFileData.disable();
          this.deployForm.patchValue({valuesYaml: templateFile});
          if (!this.deployForm.value.releaseName) {
            const releaseName = this.selectedDeploymentTarget()?.name.trim().toLowerCase().replaceAll(/\W+/g, '-');
            this.deployForm.patchValue({releaseName});
          }
        } else {
          this.deployForm.controls.envFileData.enable();
          this.deployForm.controls.envFileData.patchValue(templateFile ?? '');
          this.deployForm.controls.releaseName.disable();
          this.deployForm.controls.valuesYaml.disable();
        }
      });

    this.deployForm.controls.applicationId.statusChanges.pipe(takeUntil(this.destroyed$)).subscribe((s) => {
      if (s === 'VALID') {
        this.deployForm.controls.applicationVersionId.enable();
      } else {
        this.deployForm.controls.applicationVersionId.disable();
      }
    });

    this.licenses$.pipe(takeUntil(this.destroyed$)).subscribe((licenses) => {
      if (licenses.length > 0) {
        this.deployForm.controls.applicationLicenseId.setValue(licenses[0].id);
      } else {
        this.deployForm.controls.applicationLicenseId.reset();
      }
    });

    this.selectedLicense$
      .pipe(takeUntil(this.destroyed$))
      .subscribe((license) => this.updateAvailableApplicationVersions(license?.versions));

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
        })
      );
      this.selectedDeploymentTarget.set(created as DeploymentTarget);
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
    this.deployForm.patchValue({
      deploymentTargetId: this.selectedDeploymentTarget()!.id,
    });

    this.deployForm.markAllAsTouched();
    if (!this.deployForm.valid || this.loading) {
      return;
    }

    try {
      this.loading = true;
      const deployment = this.deployForm.value;
      if (deployment.valuesYaml) {
        deployment.valuesYaml = btoa(deployment.valuesYaml);
      } else {
        deployment.valuesYaml = undefined;
      }
      if (deployment.envFileData) {
        deployment.envFileData = btoa(deployment.envFileData);
      } else {
        deployment.envFileData = undefined;
      }
      await firstValueFrom(this.deploymentTargets.deploy(deployment as DeploymentRequest));
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

  updatedSelectedApplication(applications: Application[], applicationId?: string | null) {
    this.selectedApplication = applications.find((a) => a.id === applicationId);
  }

  updateAvailableApplicationVersions(versions: ApplicationVersion[] | undefined) {
    this.availableApplicationVersions = versions ?? [];
    if (this.availableApplicationVersions.length > 0) {
      // Only update the form control, if the previously selected version is no longer in the list
      if (
        this.availableApplicationVersions.every((version) => version.id !== this.deployForm.value.applicationVersionId)
      ) {
        this.deployForm.controls.applicationVersionId.patchValue(
          this.availableApplicationVersions[this.availableApplicationVersions.length - 1].id
        );
      }
    } else {
      this.deployForm.controls.applicationVersionId.reset();
    }
  }
}
