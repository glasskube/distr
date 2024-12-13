import {Component, EventEmitter, inject, OnDestroy, OnInit, Output, ViewChild} from '@angular/core';
import {DeploymentTargetsService} from '../../services/deployment-targets.service';
import {faShip, faXmark} from '@fortawesome/free-solid-svg-icons';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {modalFlyInOut} from '../../animations/modal';
import {CdkStep, CdkStepper} from '@angular/cdk/stepper';
import {ApplicationsService} from '../../services/applications.service';
import {firstValueFrom, Subject, takeUntil, tap} from 'rxjs';
import {DeploymentService} from '../../services/deployment.service';
import {Application} from '../../types/application';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {InstallationWizardStepperComponent} from './installation-wizard-stepper.component';
import {AsyncPipe} from '@angular/common';
import {Deployment} from '../../types/deployment';
import {ConnectInstructionsComponent} from '../../components/connect-instructions/connect-instructions.component';
import {DeploymentTargetViewModel} from '../../deployments/DeploymentTargetViewModel';

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

  private readonly applications = inject(ApplicationsService);
  private readonly deploymentTargets = inject(DeploymentTargetsService);
  private readonly deployments = inject(DeploymentService);

  @ViewChild('stepper') private stepper?: CdkStepper;

  @Output('closed') readonly closed = new EventEmitter<void>();

  readonly deploymentTargetForm = new FormGroup({
    type: new FormControl<string>('docker', Validators.required),
    name: new FormControl<string>('', Validators.required),
  });

  readonly agentForm = new FormGroup({});

  readonly deployForm = new FormGroup({
    deploymentTargetId: new FormControl<string | undefined>(undefined, Validators.required),
    applicationId: new FormControl<string | undefined>(undefined, Validators.required),
    applicationVersionId: new FormControl<string | undefined>({value: undefined, disabled: true}, Validators.required),
    notes: new FormControl<string | undefined>(undefined),
  });

  readonly applications$ = this.applications.list();
  protected selectedApplication?: Application;
  protected selectedDeploymentTarget?: DeploymentTargetViewModel;
  private loading = false;
  private readonly destroyed$ = new Subject<void>();

  ngOnInit() {
    this.deployForm.controls.applicationId.valueChanges
      .pipe(takeUntil(this.destroyed$))
      .subscribe((it) => this.updatedSelectedApplication(it!));
    this.deployForm.controls.applicationId.statusChanges.pipe(takeUntil(this.destroyed$)).subscribe((s) => {
      if (s === 'VALID') {
        this.deployForm.controls.applicationVersionId.enable();
      } else {
        this.deployForm.controls.applicationVersionId.disable();
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
        })
      );
      this.selectedDeploymentTarget = created as DeploymentTargetViewModel;
      this.nextStep();
    } else if (this.stepper?.selectedIndex === 1) {
      this.loading = true;
      this.nextStep();
    } else if (this.stepper?.selectedIndex == 2) {
      this.deployForm.patchValue({
        deploymentTargetId: this.selectedDeploymentTarget!.id,
      });

      if (!this.deployForm.valid) {
        this.deployForm.markAllAsTouched();
        return;
      }

      if (this.deployForm.valid) {
        this.loading = true;
        await this.saveDeployment();
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
      await firstValueFrom(this.deployments.create(deployment as Deployment));
      this.selectedDeploymentTarget!.latestDeployment = this.deploymentTargets.latestDeploymentFor(
        this.selectedDeploymentTarget!.id!
      );
    }
  }

  async updatedSelectedApplication(applicationId: string) {
    let applications = await firstValueFrom(this.applications$);
    this.selectedApplication = applications.find((a) => a.id === applicationId);
  }

  protected readonly faShip = faShip;
}
