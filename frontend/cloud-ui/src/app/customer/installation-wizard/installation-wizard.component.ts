import {Component, EventEmitter, inject, OnDestroy, Output, ViewChild} from '@angular/core';
import {DeploymentTargetsService} from '../../services/deployment-targets.service';
import {faShip, faXmark} from '@fortawesome/free-solid-svg-icons';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {modalFlyInOut} from '../../animations/modal';
import {CdkStep, CdkStepper} from '@angular/cdk/stepper';
import {ApplicationsService} from '../../services/applications.service';
import {firstValueFrom, tap} from 'rxjs';
import {DeploymentService} from '../../services/deployment.service';
import {Application} from '../../types/application';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {InstallationWizardStepperComponent} from './installation-wizard-stepper.component';
import {AsyncPipe} from '@angular/common';
import {Deployment} from '../../types/deployment';
import {DeploymentTargetViewModel} from '../../deployment-targets/DeploymentTargetViewModel';
import {ConnectInstructionsComponent} from '../../components/connect-instructions/connect-instructions.component';

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
export class InstallationWizardComponent implements OnDestroy {
  protected readonly xmarkIcon = faXmark;
  private applications = inject(ApplicationsService);
  private deploymentTargets = inject(DeploymentTargetsService);
  private deployments = inject(DeploymentService);
  @ViewChild('stepper') stepper!: CdkStepper;

  @Output('closed') closed = new EventEmitter<void>();

  deploymentTargetForm = new FormGroup({
    type: new FormControl<string>('docker', Validators.required),
    name: new FormControl<string>('', Validators.required),
  });

  agentForm = new FormGroup({});

  private loading = false;

  ngOnDestroy() {
    this.applicationIdChange$.unsubscribe();
  }

  async attemptContinue() {
    if (this.loading) {
      return;
    }

    if (this.stepper.selectedIndex === 0) {
      if (!this.deploymentTargetForm.valid) {
        this.deploymentTargetForm.markAllAsTouched();
        return;
      }

      this.deploymentTargets
        .create({
          name: this.deploymentTargetForm.value.name!,
          type: this.deploymentTargetForm.value.type!,
        })
        .pipe(tap((dt) => (this.selectedDeploymentTarget = dt as DeploymentTargetViewModel)))
        .subscribe(this.nextStep);

      this.loading = true;
      this.nextStep();
    } else if (this.stepper.selectedIndex === 1) {
      this.loading = true;
      this.nextStep();
    } else if (this.stepper.selectedIndex == 2) {
      this.deployForm.patchValue({
        deploymentTargetId: this.selectedDeploymentTarget!!.id,
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
    this.stepper.next();
  }

  protected readonly shipIcon = faShip;

  applications$ = this.applications.list();
  selectedApplication?: Application | null;

  deployForm = new FormGroup({
    deploymentTargetId: new FormControl<string | undefined>(undefined, Validators.required),
    applicationId: new FormControl<string | undefined>(undefined, Validators.required),
    applicationVersionId: new FormControl<string | undefined>(undefined, Validators.required),
    notes: new FormControl<string | undefined>(undefined),
  });

  protected selectedDeploymentTarget?: DeploymentTargetViewModel | null;

  private readonly applicationIdChange$ = this.deployForm.controls.applicationId.valueChanges.subscribe((it) =>
    this.updatedSelectedApplication(it!!)
  );

  async saveDeployment() {
    if (this.deployForm.valid) {
      const deployment = this.deployForm.value;
      await firstValueFrom(this.deployments.create(deployment as Deployment));
      this.selectedDeploymentTarget!!.latestDeployment = this.deploymentTargets.latestDeploymentFor(
        this.selectedDeploymentTarget!!.id!!
      );
    }
  }

  async updatedSelectedApplication(applicationId: string) {
    let applications = await firstValueFrom(this.applications$);
    this.selectedApplication = applications.find((a) => a.id === applicationId) || null;
  }
}
