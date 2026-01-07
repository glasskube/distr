import {GlobalPositionStrategy, OverlayModule} from '@angular/cdk/overlay';
import {DatePipe, NgOptimizedImage, NgTemplateOutlet} from '@angular/common';
import {
  Component,
  computed,
  inject,
  input,
  resource,
  signal,
  TemplateRef,
  ViewChild,
  WritableSignal,
} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {
  faCircleExclamation,
  faEllipsisVertical,
  faHeartPulse,
  faLink,
  faPen,
  faPlus,
  faRotate,
  faShip,
  faTrash,
  faTriangleExclamation,
  faXmark,
} from '@fortawesome/free-solid-svg-icons';
import {
  DeploymentTarget,
  DeploymentTargetScope,
  DeploymentType,
  DeploymentWithLatestRevision,
} from '@glasskube/distr-sdk';
import {filter, firstValueFrom, lastValueFrom, switchMap} from 'rxjs';
import {SemVer} from 'semver';
import {getFormDisplayedError} from '../../../util/errors';
import {IsStalePipe} from '../../../util/model';
import {drawerFlyInOut} from '../../animations/drawer';
import {dropdownAnimation} from '../../animations/dropdown';
import {modalFlyInOut} from '../../animations/modal';
import {ConnectInstructionsComponent} from '../../components/connect-instructions/connect-instructions.component';
import {DeploymentStatusDot, StatusDotComponent} from '../../components/status-dot';
import {UuidComponent} from '../../components/uuid';
import {AgentVersionService} from '../../services/agent-version.service';
import {AuthService} from '../../services/auth.service';
import {DeploymentTargetLatestMetrics} from '../../services/deployment-target-metrics.service';
import {DeploymentTargetsService} from '../../services/deployment-targets.service';
import {DialogRef, OverlayService} from '../../services/overlay.service';
import {ToastService} from '../../services/toast.service';
import {DeploymentModalComponent} from '../deployment-modal.component';
import {DeploymentStatusModalComponent} from '../deployment-status-modal/deployment-status-modal.component';
import {DeploymentTargetMetricsComponent} from './deployment-target-metrics.component';

@Component({
  selector: 'app-deployment-target-card',
  templateUrl: './deployment-target-card.component.html',
  imports: [
    NgOptimizedImage,
    StatusDotComponent,
    UuidComponent,
    DatePipe,
    FaIconComponent,
    IsStalePipe,
    DeploymentStatusDot,
    OverlayModule,
    ConnectInstructionsComponent,
    ReactiveFormsModule,
    DeploymentModalComponent,
    DeploymentTargetMetricsComponent,
    NgTemplateOutlet,
    DeploymentStatusModalComponent,
  ],
  animations: [modalFlyInOut, drawerFlyInOut, dropdownAnimation],
})
export class DeploymentTargetCardComponent {
  private readonly agentVersionsSvc = inject(AgentVersionService);
  private readonly overlay = inject(OverlayService);
  protected readonly auth = inject(AuthService);
  private readonly deploymentTargets = inject(DeploymentTargetsService);
  private readonly toast = inject(ToastService);

  protected readonly customerManagedWarning = `
    You are about to make changes to a customer-managed deployment.
    Ensure this is done in coordination with the customer.`;

  public readonly deploymentTarget = input.required<DeploymentTarget>();
  public readonly fullVersion = input(true);
  public readonly deploymentTargetMetrics = input<DeploymentTargetLatestMetrics | undefined>(undefined);

  @ViewChild('deploymentModal') protected readonly deploymentModal!: TemplateRef<unknown>;
  @ViewChild('deploymentStatusModal') protected readonly deploymentStatusModal!: TemplateRef<unknown>;
  @ViewChild('instructionsModal') protected readonly instructionsModal!: TemplateRef<unknown>;
  @ViewChild('deleteConfirmModal') protected readonly deleteConfirmModal!: TemplateRef<unknown>;
  @ViewChild('manageDeploymentTargetDrawer') protected readonly manageDeploymentTargetDrawer!: TemplateRef<unknown>;
  @ViewChild('deleteDeploymentProgressModal') protected readonly deleteDeploymentProgressModal!: TemplateRef<unknown>;

  protected readonly faShip = faShip;
  protected readonly faLink = faLink;
  protected readonly faEllipsisVertical = faEllipsisVertical;
  protected readonly faPen = faPen;
  protected readonly faTrash = faTrash;
  protected readonly faHeartPulse = faHeartPulse;
  protected readonly faTriangleExclamation = faTriangleExclamation;
  protected readonly faXmark = faXmark;
  protected readonly faCircleExclamation = faCircleExclamation;
  protected readonly faRotate = faRotate;

  protected readonly showDeploymentTargetDropdown = signal(false);
  protected readonly showDeploymentDropdownForId = signal<string | undefined>(undefined);
  protected readonly selectedDeploymentTarget = signal<DeploymentTarget | undefined>(undefined);
  protected readonly selectedDeployment = signal<DeploymentWithLatestRevision | undefined>(undefined);

  protected readonly metricsOpened = signal(false);

  protected readonly agentVersions = resource({
    loader: () => firstValueFrom(this.agentVersionsSvc.list()),
  });

  protected readonly isUndeploySupported = this.isAgentVersionAtLeast('1.3.0');
  protected readonly isMultiDeploymentSupported = this.isAgentVersionAtLeast('1.6.0');
  protected readonly isLoggingSupported = this.isAgentVersionAtLeast('1.9.0');
  protected readonly isForceRestartSupported = this.isAgentVersionAtLeast('1.12.0');

  protected readonly agentUpdateAvailable = computed(() => {
    const agentVersions = this.agentVersions.value() ?? [];
    return (
      agentVersions.length > 0 &&
      this.deploymentTarget().agentVersion?.id !== agentVersions[agentVersions.length - 1].id
    );
  });

  protected readonly agentUpdatePending = computed(
    () =>
      this.deploymentTarget().currentStatus !== undefined &&
      this.deploymentTarget().agentVersion?.id !== this.deploymentTarget().reportedAgentVersionId
  );

  protected readonly editForm = new FormGroup({
    id: new FormControl<string | undefined>(undefined),
    name: new FormControl('', Validators.required),
    type: new FormControl<DeploymentType | undefined>({value: undefined, disabled: true}, Validators.required),
    namespace: new FormControl<string | undefined>({value: undefined, disabled: true}),
    scope: new FormControl<DeploymentTargetScope>({value: 'namespace', disabled: true}),
    metricsEnabled: new FormControl<boolean>(true),
  });
  protected editFormLoading = false;

  private modal?: DialogRef;
  private manageDeploymentTargetRef?: DialogRef;

  protected async showDeploymentModal(deployment?: DeploymentWithLatestRevision) {
    this.selectedDeploymentTarget.set(this.deploymentTarget());
    this.selectedDeployment.set(deployment);
    this.showModal(this.deploymentModal);
  }

  protected async saveDeploymentTarget() {
    this.editForm.markAllAsTouched();
    if (this.editForm.valid) {
      this.editFormLoading = true;
      const val = this.editForm.value;
      const dt: DeploymentTarget = {
        id: val.id!,
        name: val.name!,
        type: val.type!,
        deployments: [],
        metricsEnabled: val.metricsEnabled ?? false,
      };

      try {
        this.loadDeploymentTarget(
          await lastValueFrom(val.id ? this.deploymentTargets.update(dt) : this.deploymentTargets.create(dt))
        );
        this.toast.success(`${dt.name} saved successfully`);
        this.hideDrawer();
      } catch (e) {
        const msg = getFormDisplayedError(e);
        if (msg) {
          this.toast.error(msg);
        }
      } finally {
        this.editFormLoading = false;
      }
    }
  }

  private loadDeploymentTarget(dt: DeploymentTarget) {
    this.editForm.patchValue({
      ...dt,
    });
    if (dt.scope === 'namespace') {
      this.editForm.controls.metricsEnabled.disable();
    } else {
      this.editForm.controls.metricsEnabled.enable();
    }
  }

  protected async openInstructionsModal() {
    const dt = this.deploymentTarget();
    if (dt.currentStatus !== undefined) {
      const message = `If you continue, the previous authentication secret for ${dt.name} becomes invalid. Continue?`;
      const alert =
        dt.customerOrganization !== undefined && this.auth.isVendor()
          ? ({type: 'warning', message: this.customerManagedWarning} as const)
          : undefined;
      if (!(await firstValueFrom(this.overlay.confirm({message: {message, alert}})))) {
        return;
      }
    }
    this.showModal(this.instructionsModal);
  }

  protected openStatusModal(deployment: DeploymentWithLatestRevision) {
    if (deployment?.id) {
      this.selectedDeployment.set(deployment);
      this.showModal(this.deploymentStatusModal);
    }
  }

  protected setLogsEnabled(deplyoment: DeploymentWithLatestRevision, logsEnabled: boolean) {
    if (deplyoment.id) {
      this.deploymentTargets.patchDeployment(deplyoment.id, {logsEnabled}).subscribe({
        next: () => this.toast.success('Deployment has been updated.'),
        error: (e) => {
          const msg = getFormDisplayedError(e);
          if (msg) {
            this.toast.error(msg);
          }
        },
      });
    }
  }

  protected forceRestart(deployment: DeploymentWithLatestRevision) {
    if (deployment.id) {
      this.overlay
        .confirm({
          message: {
            message: 'Are you sure you want to force restart this deployment?',
            alert: {
              type: 'warning',
              message: 'Depending on the deployment, this may cause downtime.',
            },
          },
        })
        .pipe(
          filter((result) => result === true),
          switchMap(() =>
            this.deploymentTargets.deploy({
              deploymentId: deployment.id,
              deploymentTargetId: deployment.deploymentTargetId,
              applicationVersionId: deployment.applicationVersionId,
              applicationLicenseId: deployment.applicationLicenseId,
              releaseName: deployment.releaseName,
              dockerType: deployment.dockerType,
              valuesYaml: deployment.valuesYaml,
              envFileData: deployment.envFileData,
              forceRestart: true,
            })
          )
        )
        .subscribe({
          next: () => this.toast.success('Deployment has been restarted.'),
          error: (e) => {
            const msg = getFormDisplayedError(e);
            if (msg) {
              this.toast.error(msg);
            }
          },
        });
    }
  }

  protected deleteDeploymentTarget() {
    const dt = this.deploymentTarget();
    const alert =
      dt.customerOrganization !== undefined && this.auth.isVendor()
        ? ({type: 'warning', message: this.customerManagedWarning} as const)
        : undefined;
    this.overlay
      .confirm({
        customTemplate: this.deleteConfirmModal,
        requiredConfirmInputText: 'DELETE',
        message: {
          alert,
          message: '',
        },
      })
      .pipe(
        filter((result) => result === true),
        switchMap(() => this.deploymentTargets.delete(dt))
      )
      .subscribe({
        error: (e) => {
          const msg = getFormDisplayedError(e);
          if (msg) {
            this.toast.error(msg);
          }
        },
      });
  }

  protected async deleteDeployment(d: DeploymentWithLatestRevision, confirmTemplate: TemplateRef<any>) {
    const dt = this.deploymentTarget();
    const alert =
      dt.customerOrganization !== undefined && this.auth.isVendor()
        ? ({type: 'warning', message: this.customerManagedWarning} as const)
        : undefined;
    if (d.id) {
      if (
        await firstValueFrom(
          this.overlay.confirm({
            customTemplate: confirmTemplate,
            requiredConfirmInputText: 'UNDEPLOY',
            message: {
              alert,
              message: '',
            },
          })
        )
      ) {
        const modalRef = this.overlay.showModal(this.deleteDeploymentProgressModal, {
          positionStrategy: new GlobalPositionStrategy().centerHorizontally().centerVertically(),
          backdropStyleOnly: true,
        });

        try {
          await firstValueFrom(this.deploymentTargets.undeploy(d.id));
        } catch (e) {
          const msg = getFormDisplayedError(e);
          if (msg) {
            this.toast.error(msg);
          }
        } finally {
          modalRef?.dismiss();
        }
      }
    }
  }

  public async updateDeploymentTargetAgent(): Promise<void> {
    try {
      const dt = this.deploymentTarget();
      const agentVersions = this.agentVersions.value();
      if (agentVersions?.length) {
        const targetVersion = agentVersions[agentVersions.length - 1];
        if (
          await firstValueFrom(
            this.overlay.confirm(`Update ${dt.name} agent from ${dt.agentVersion?.name} to ${targetVersion.name}?`)
          )
        ) {
          dt.agentVersion = targetVersion;
          await firstValueFrom(this.deploymentTargets.update(dt));
        }
      }
    } catch (e) {}
  }

  protected showModal(templateRef: TemplateRef<unknown>) {
    this.hideModal();
    this.modal = this.overlay.showModal(templateRef, {
      positionStrategy: new GlobalPositionStrategy().centerHorizontally().centerVertically(),
    });
  }

  protected hideModal(): void {
    this.modal?.close();
  }

  protected openDrawer() {
    this.hideDrawer();
    this.loadDeploymentTarget(this.deploymentTarget());
    this.manageDeploymentTargetRef = this.overlay.showDrawer(this.manageDeploymentTargetDrawer);
  }

  protected hideDrawer() {
    this.manageDeploymentTargetRef?.close();
    this.resetEditForm();
  }

  private resetEditForm() {
    this.editForm.reset();
    this.editForm.patchValue({type: 'docker'});
  }

  private isAgentVersionAtLeast(version: string, allowSnapshot = true) {
    return computed(() => {
      if (!this.deploymentTarget().reportedAgentVersionId) {
        console.warn('reported agent version id is empty');
        return true;
      }
      const reported = this.agentVersions
        .value()
        ?.find((it) => it.id === this.deploymentTarget().reportedAgentVersionId);
      if (!reported) {
        console.warn('agent version with id not found', this.deploymentTarget().reportedAgentVersionId);
        return false;
      }
      try {
        return (allowSnapshot && reported.name === 'snapshot') || new SemVer(reported.name).compare(version) >= 0;
      } catch (e) {
        console.warn(e);
        return allowSnapshot && reported.name === 'snapshot';
      }
    });
  }

  protected toggle(signal: WritableSignal<boolean>) {
    signal.update((val) => !val);
  }

  protected readonly faPlus = faPlus;
}
