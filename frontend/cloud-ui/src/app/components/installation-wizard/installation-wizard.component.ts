import {CdkStep, CdkStepper} from '@angular/cdk/stepper';
import {AsyncPipe} from '@angular/common';
import {Component, EventEmitter, inject, OnDestroy, OnInit, Output, signal, ViewChild} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faShip, faXmark} from '@fortawesome/free-solid-svg-icons';
import {combineLatest, firstValueFrom, map, of, Subject, switchMap, takeUntil, tap, withLatestFrom} from 'rxjs';
import {modalFlyInOut} from '../../animations/modal';
import {ConnectInstructionsComponent} from '../../components/connect-instructions/connect-instructions.component';
import {DeploymentTargetViewModel} from '../../deployments/deployment-target-view-model';
import {ApplicationsService} from '../../services/applications.service';
import {DeploymentTargetsService} from '../../services/deployment-targets.service';
import {DeploymentService} from '../../services/deployment.service';
import {ToastService} from '../../services/toast.service';
import {Application} from '../../types/application';
import {Deployment} from '../../types/deployment';
import {InstallationWizardStepperComponent} from './installation-wizard-stepper.component';
import {toObservable} from '@angular/core/rxjs-interop';

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
    type: new FormControl('docker', Validators.required),
    name: new FormControl('', Validators.required),
    namespace: new FormControl({value: '', disabled: true}, Validators.required),
  });

  readonly agentForm = new FormGroup({});

  readonly deployForm = new FormGroup({
    deploymentTargetId: new FormControl<string | undefined>(undefined, Validators.required),
    applicationId: new FormControl<string | undefined>(undefined, Validators.required),
    applicationVersionId: new FormControl<string | undefined>({value: undefined, disabled: true}, Validators.required),
    valuesYaml: new FormControl<string | undefined>({value: undefined, disabled: true}),
    notes: new FormControl<string | undefined>(undefined),
  });

  protected selectedApplication?: Application;
  protected selectedDeploymentTarget = signal<DeploymentTargetViewModel | null>(null);
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
          this.deployForm.controls.valuesYaml.enable();
          this.deployForm.patchValue({valuesYaml});
        } else {
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
      } else {
        this.deploymentTargetForm.controls.namespace.disable();
      }
    });
  }

  ngOnDestroy() {
    this.destroyed$.next();
  }

  async attemptContinue() {
    if (this.loading) {
      return;
    }

    if (this.stepper?.selectedIndex === 0) {
      if (!this.deploymentTargetForm.valid) {
        this.deploymentTargetForm.markAllAsTouched();
        return;
      }

      this.loading = true;
      const created = await firstValueFrom(
        this.deploymentTargets.create({
          name: this.deploymentTargetForm.value.name!,
          type: this.deploymentTargetForm.value.type!,
          namespace: this.deploymentTargetForm.value.namespace!,
        })
      );
      this.selectedDeploymentTarget.set(created as DeploymentTargetViewModel);
      this.nextStep();
    } else if (this.stepper?.selectedIndex === 1) {
      this.loading = true;
      this.nextStep();
    } else if (this.stepper?.selectedIndex == 2) {
      this.deployForm.patchValue({
        deploymentTargetId: this.selectedDeploymentTarget()!.id,
      });

      if (!this.deployForm.valid) {
        this.deployForm.markAllAsTouched();
        return;
      }

      if (this.deployForm.valid) {
        this.loading = true;
        await this.saveDeployment();
        this.toast.success('Deployment saved successfully');
        this.close();
      }
    }
  }

  close() {
    this.closed.emit();
  }

  private nextStep() {
    this.loading = false;
    this.stepper?.next();
  }

  async saveDeployment() {
    if (this.deployForm.valid) {
      const deployment = this.deployForm.value;
      if (deployment.valuesYaml) {
        deployment.valuesYaml = btoa(deployment.valuesYaml);
      }
      await firstValueFrom(this.deployments.create(deployment as Deployment));
      this.selectedDeploymentTarget()!.latestDeployment = this.deploymentTargets.latestDeploymentFor(
        this.selectedDeploymentTarget()!.id!
      );
    }
  }

  updatedSelectedApplication(applications: Application[], applicationId?: string | null) {
    this.selectedApplication = applications.find((a) => a.id === applicationId);
  }
}
