import {CdkStep, CdkStepper} from '@angular/cdk/stepper';
import {AsyncPipe} from '@angular/common';
import {Component, EventEmitter, inject, OnDestroy, OnInit, Output, signal, ViewChild} from '@angular/core';
import {toObservable} from '@angular/core/rxjs-interop';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faShip, faXmark} from '@fortawesome/free-solid-svg-icons';
import {combineLatest, firstValueFrom, map, of, Subject, switchMap, takeUntil, tap, withLatestFrom} from 'rxjs';
import {getFormDisplayedError} from '../../../util/errors';
import {modalFlyInOut} from '../../animations/modal';
import {ConnectInstructionsComponent} from '../../components/connect-instructions/connect-instructions.component';
import {ApplicationsService} from '../../services/applications.service';
import {DeploymentTargetsService} from '../../services/deployment-targets.service';
import {DeploymentService} from '../../services/deployment.service';
import {ToastService} from '../../services/toast.service';
import {Application} from '../../types/application';
import {DeploymentRequest, DeploymentTargetScope, DeploymentType} from '../../types/deployment';
import {DeploymentTarget} from '../../types/deployment-target';
import {YamlEditorComponent} from '../yaml-editor.component';
import {InstallationWizardStepperComponent} from './installation-wizard-stepper.component';
import {AutotrimDirective} from '../../directives/autotrim.directive';

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
  private readonly deployments = inject(DeploymentService);

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
    valuesYaml: new FormControl<string>({value: '', disabled: true}),
    releaseName: new FormControl<string>({value: '', disabled: true}, Validators.required),
    notes: new FormControl<string | undefined>(undefined),
  });

  protected selectedApplication?: Application;
  protected selectedDeploymentTarget = signal<DeploymentTarget | null>(null);
  readonly applications$ = combineLatest([this.applications.list(), toObservable(this.selectedDeploymentTarget)]).pipe(
    map(([apps, dt]) => apps.filter((app) => app.type === dt?.type))
  );
  private loading = false;
  private readonly destroyed$ = new Subject<void>();

  ngOnInit() {
    this.deployForm.controls.applicationId.valueChanges
      .pipe(
        takeUntil(this.destroyed$),
        withLatestFrom(this.applications$),
        tap(([selected, apps]) => this.updatedSelectedApplication(apps, selected))
      )
      .subscribe(() => {
        if (this.selectedApplication && (this.selectedApplication.versions ?? []).length > 0) {
          const versions = this.selectedApplication.versions!;
          this.deployForm.controls.applicationVersionId.patchValue(versions[versions.length - 1].id);
        } else {
          this.deployForm.controls.applicationVersionId.reset();
        }
      });
    this.deployForm.controls.applicationVersionId.valueChanges
      .pipe(
        takeUntil(this.destroyed$),
        switchMap((id) =>
          this.selectedApplication?.type === 'kubernetes'
            ? this.applications
                .getTemplateFile(this.selectedApplication?.id!, id!)
                .pipe(map((data) => [this.selectedApplication?.type, data]))
            : of([this.selectedApplication?.type, null])
        )
      )
      .subscribe(([type, valuesYaml]) => {
        if (type === 'kubernetes') {
          this.deployForm.controls.releaseName.enable();
          this.deployForm.controls.valuesYaml.enable();
          this.deployForm.patchValue({valuesYaml});
          if (!this.deployForm.value.releaseName) {
            const releaseName = this.selectedDeploymentTarget()?.name.trim().toLowerCase().replaceAll(/\W+/g, '-');
            this.deployForm.patchValue({releaseName});
          }
        } else {
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
      await firstValueFrom(this.deployments.createOrUpdate(deployment as DeploymentRequest));
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
}
