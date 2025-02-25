import {GlobalPositionStrategy, OverlayModule} from '@angular/cdk/overlay';
import {AsyncPipe, DatePipe, NgOptimizedImage, UpperCasePipe} from '@angular/common';
import {AfterViewInit, Component, inject, Input, OnDestroy, signal, TemplateRef, ViewChild} from '@angular/core';
import {toObservable} from '@angular/core/rxjs-interop';
import {FormControl, FormGroup, FormsModule, ReactiveFormsModule, Validators} from '@angular/forms';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {
  faCircleExclamation,
  faEllipsisVertical,
  faHeartPulse,
  faLightbulb,
  faMagnifyingGlass,
  faPen,
  faPlus,
  faShip,
  faTrash,
  faTriangleExclamation,
  faXmark,
} from '@fortawesome/free-solid-svg-icons';
import {
  Application,
  ApplicationVersion,
  Deployment,
  DeploymentRequest,
  DeploymentRevisionStatus,
  DeploymentTarget,
  DeploymentTargetScope,
  DeploymentType,
  DeploymentWithLatestRevision,
} from '@glasskube/distr-sdk';
import {
  catchError,
  combineLatest,
  EMPTY,
  filter,
  first,
  firstValueFrom,
  lastValueFrom,
  map,
  Observable,
  of,
  Subject,
  switchMap,
} from 'rxjs';
import {SemVer} from 'semver';
import {maxBy} from '../../util/arrays';
import {isArchived} from '../../util/dates';
import {getFormDisplayedError} from '../../util/errors';
import {filteredByFormControl} from '../../util/filter';
import {IsStalePipe} from '../../util/model';
import {drawerFlyInOut} from '../animations/drawer';
import {dropdownAnimation} from '../animations/dropdown';
import {modalFlyInOut} from '../animations/modal';
import {ConnectInstructionsComponent} from '../components/connect-instructions/connect-instructions.component';
import {InstallationWizardComponent} from '../components/installation-wizard/installation-wizard.component';
import {StatusDotComponent} from '../components/status-dot';
import {UuidComponent} from '../components/uuid';
import {DeploymentFormComponent, DeploymentFormValue} from '../deployment-form/deployment-form.component';
import {AutotrimDirective} from '../directives/autotrim.directive';
import {AgentVersionService} from '../services/agent-version.service';
import {ApplicationsService} from '../services/applications.service';
import {AuthService} from '../services/auth.service';
import {DeploymentStatusService} from '../services/deployment-status.service';
import {DeploymentTargetsService} from '../services/deployment-targets.service';
import {LicensesService} from '../services/licenses.service';
import {DialogRef, OverlayService} from '../services/overlay.service';
import {ToastService} from '../services/toast.service';

@Component({
  selector: 'app-deployment-targets',
  imports: [
    AsyncPipe,
    DatePipe,
    FaIconComponent,
    FormsModule,
    ReactiveFormsModule,
    NgOptimizedImage,
    IsStalePipe,
    StatusDotComponent,
    ConnectInstructionsComponent,
    InstallationWizardComponent,
    UpperCasePipe,
    AutotrimDirective,
    UuidComponent,
    DeploymentFormComponent,
    OverlayModule,
  ],
  templateUrl: './deployment-targets.component.html',
  standalone: true,
  animations: [modalFlyInOut, drawerFlyInOut, dropdownAnimation],
})
export class DeploymentTargetsComponent implements AfterViewInit, OnDestroy {
  @Input('fullVersion') fullVersion = false;

  public readonly auth = inject(AuthService);
  private readonly toast = inject(ToastService);
  private readonly overlay = inject(OverlayService);
  private readonly applications = inject(ApplicationsService);
  private readonly licenses = inject(LicensesService);
  private readonly deploymentTargets = inject(DeploymentTargetsService);
  private readonly deploymentStatuses = inject(DeploymentStatusService);
  private readonly agentVersions = inject(AgentVersionService);

  readonly magnifyingGlassIcon = faMagnifyingGlass;
  readonly plusIcon = faPlus;
  readonly penIcon = faPen;
  readonly shipIcon = faShip;
  readonly xmarkIcon = faXmark;
  protected readonly faHeartPulse = faHeartPulse;
  protected readonly faTrash = faTrash;
  protected readonly faEllipsisVertical = faEllipsisVertical;
  protected readonly faTriangleExclamation = faTriangleExclamation;
  protected readonly faCircleExclamation = faCircleExclamation;
  protected readonly faLightbulb = faLightbulb;

  private destroyed$ = new Subject<void>();
  private modal?: DialogRef;
  private manageDeploymentTargetRef?: DialogRef;

  private deploymentWizardOverlayRef?: DialogRef;

  protected readonly customerManagedWarning = `
    You are about to make changes to a customer-managed deployment.
    Ensure this is done in coordination with the customer.`;

  @ViewChild('deploymentWizard') wizardRef?: TemplateRef<unknown>;
  selectedDeploymentTarget = signal<DeploymentTarget | null>(null);

  selectedApplication?: Application | null;

  readonly filterForm = new FormGroup({
    search: new FormControl(''),
  });
  readonly editForm = new FormGroup({
    id: new FormControl<string | undefined>(undefined),
    name: new FormControl('', Validators.required),
    type: new FormControl<DeploymentType | undefined>({value: undefined, disabled: true}, Validators.required),
    geolocation: new FormGroup({
      lat: new FormControl<number | undefined>(undefined),
      lon: new FormControl<number | undefined>(undefined),
    }),
    namespace: new FormControl<string | undefined>({value: undefined, disabled: true}),
    scope: new FormControl<DeploymentTargetScope>({value: 'namespace', disabled: true}),
  });
  editFormLoading = false;

  readonly deployForm = new FormControl<DeploymentFormValue | undefined>(undefined, Validators.required);
  deployFormLoading = false;

  readonly deploymentTargets$ = this.deploymentTargets.poll();

  readonly filteredDeploymentTargets$ = filteredByFormControl(
    this.deploymentTargets$,
    this.filterForm.controls.search,
    (dt, search) => !search || (dt.name || '').toLowerCase().includes(search.toLowerCase())
  );
  private readonly applications$ = this.applications.list();
  public readonly agentVersions$ = this.agentVersions.list();

  readonly showAgentUpdateColumn$ = combineLatest([this.filteredDeploymentTargets$, this.agentVersions$]).pipe(
    map(
      ([dts, avs]) =>
        avs.length !== 0 && dts.some((dt) => dt.agentVersion?.id && dt.agentVersion?.id !== avs[avs.length - 1].id)
    )
  );

  readonly filteredApplications$ = combineLatest([
    this.applications$,
    toObservable(this.selectedDeploymentTarget),
  ]).pipe(map(([apps, dt]) => apps.filter((app) => app.type === dt?.type)));

  statuses: Observable<DeploymentRevisionStatus[]> = EMPTY;

  protected deploymentTargetsWithUpdate$: Observable<{dt: DeploymentTarget; version: ApplicationVersion}[]> =
    this.deploymentTargets$.pipe(
      switchMap((deploymentTargets) =>
        combineLatest(
          deploymentTargets
            .filter((deplyomentTarget) => deplyomentTarget.id && deplyomentTarget.deployment)
            .map((deploymentTarget) =>
              this.getAvailableVersions(deploymentTarget.deployment!).pipe(
                map((versions) => ({dt: deploymentTarget, version: this.findMaxVersion(versions)}))
              )
            )
        )
      ),
      map((result) =>
        result
          .filter((it) => it.version && it.version.id !== it.dt.deployment?.applicationVersionId)
          .map((it) => ({dt: it.dt, version: it.version!}))
      )
    );

  protected showDropdownForId?: string;

  ngAfterViewInit() {
    if (this.fullVersion) {
      combineLatest([this.applications$, this.deploymentTargets$])
        .pipe(first())
        .subscribe(([apps, dts]) => {
          if (this.auth.hasRole('customer') && apps.length > 0 && dts.length === 0) {
            this.openWizard();
          }
        });
    }
  }

  ngOnDestroy(): void {
    this.destroyed$.next();
    this.destroyed$.complete();
  }

  openDrawer(templateRef: TemplateRef<unknown>, deploymentTarget: DeploymentTarget) {
    this.hideDrawer();
    this.selectedDeploymentTarget.set(deploymentTarget);
    this.loadDeploymentTarget(deploymentTarget);
    this.manageDeploymentTargetRef = this.overlay.showDrawer(templateRef);
  }

  hideDrawer() {
    this.manageDeploymentTargetRef?.close();
    this.resetEditForm();
  }

  openWizard() {
    this.deploymentWizardOverlayRef?.close();
    this.deploymentWizardOverlayRef = this.overlay.showModal(this.wizardRef!, {
      hasBackdrop: true,
      backdropStyleOnly: true,
      positionStrategy: new GlobalPositionStrategy().centerHorizontally().centerVertically(),
    });
  }

  closeWizard() {
    this.deploymentWizardOverlayRef?.close();
  }

  resetEditForm() {
    this.editForm.reset();
    this.editForm.patchValue({type: 'docker'});
  }

  async deleteDeploymentTarget(dt: DeploymentTarget) {
    const message = `
      You are about to delete the deployment with name ${dt.name}?
      Afterwards you will not be able to deploy to this target anymore.
      This will also delete all associated configuration, revision history and status logs.
      This does not undeploy the deployed application.
      This action can not be undone.
      Do you want to continue?`;
    const warning =
      dt.createdBy?.userRole === 'customer' && this.auth.hasRole('vendor')
        ? {message: this.customerManagedWarning}
        : undefined;
    this.overlay
      .confirm({message, warning})
      .pipe(
        filter((result) => result === true),
        switchMap(() => this.deploymentTargets.delete(dt)),
        catchError((e) => {
          const msg = getFormDisplayedError(e);
          if (msg) {
            this.toast.error(msg);
          }
          return EMPTY;
        })
      )
      .subscribe();
  }

  async deleteDeployment(dt: DeploymentTarget, d: Deployment) {
    const message = `
      You are about to uninstall the application installed on ${dt.name}.
      Afterwards you will be able to deploy a new application to this target.
      This will also delete all associated configuration, revision history and status logs.
      In most cases all application data is deleted.
      This action can not be undone.
      Do you want to continue?`;
    const warning =
      dt.createdBy?.userRole === 'customer' && this.auth.hasRole('vendor')
        ? {message: this.customerManagedWarning}
        : undefined;
    if (d.id) {
      if (await firstValueFrom(this.overlay.confirm({message, warning})))
        try {
          await firstValueFrom(this.deploymentTargets.undeploy(d.id));
        } catch (e) {
          const msg = getFormDisplayedError(e);
          if (msg) {
            this.toast.error(msg);
          }
        }
    }
  }

  loadDeploymentTarget(dt: DeploymentTarget) {
    this.editForm.patchValue({
      // to reset the geolocation inputs in case dt has no geolocation
      geolocation: {lat: undefined, lon: undefined},
      ...dt,
    });
  }

  showModal(templateRef: TemplateRef<unknown>) {
    this.hideModal();
    this.modal = this.overlay.showModal(templateRef, {
      positionStrategy: new GlobalPositionStrategy().centerHorizontally().centerVertically(),
    });
  }

  hideModal(): void {
    this.modal?.close();
  }

  async saveDeploymentTarget() {
    this.editForm.markAllAsTouched();
    if (this.editForm.valid) {
      this.editFormLoading = true;
      const val = this.editForm.value;
      const dt: DeploymentTarget = {
        id: val.id!,
        name: val.name!,
        type: val.type!,
      };

      if (typeof val.geolocation?.lat === 'number' && typeof val.geolocation.lon === 'number') {
        dt.geolocation = {
          lat: val.geolocation.lat,
          lon: val.geolocation.lon,
        };
      }

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

  async newDeployment(
    deploymentTarget: DeploymentTarget,
    modalTemplate: TemplateRef<any>,
    version?: ApplicationVersion
  ) {
    const apps = await firstValueFrom(this.applications$);
    if (deploymentTarget.deployment) {
      this.updatedSelectedApplication(apps, deploymentTarget.deployment.applicationId);
    }
    this.selectedDeploymentTarget.set(deploymentTarget);

    this.deployForm.reset({
      deploymentTargetId: deploymentTarget.id,
      applicationId: deploymentTarget.deployment?.applicationId,
      applicationVersionId: version?.id ?? deploymentTarget.deployment?.applicationVersionId,
      applicationLicenseId: deploymentTarget.deployment?.applicationLicenseId,
      releaseName: deploymentTarget.deployment?.releaseName,
      valuesYaml: deploymentTarget.deployment?.valuesYaml ? atob(deploymentTarget.deployment.valuesYaml) : undefined,
      envFileData: deploymentTarget.deployment?.envFileData ? atob(deploymentTarget.deployment.envFileData) : undefined,
    });

    this.showModal(modalTemplate);
  }

  async saveDeployment() {
    this.deployForm.markAllAsTouched();
    if (this.deployForm.valid) {
      this.deployFormLoading = true;
      const deployment: DeploymentRequest = {
        deploymentId: this.selectedDeploymentTarget()?.deployment?.id,
        ...(this.deployForm.value as Required<DeploymentFormValue>),
      };
      if (deployment.valuesYaml) {
        deployment.valuesYaml = btoa(deployment.valuesYaml);
      }
      if (deployment.envFileData) {
        deployment.envFileData = btoa(deployment.envFileData);
      }
      try {
        await firstValueFrom(this.deploymentTargets.deploy(deployment as DeploymentRequest));
        this.toast.success('Deployment saved successfully');
        this.hideModal();
      } catch (e) {
        const msg = getFormDisplayedError(e);
        if (msg) {
          this.toast.error(msg);
        }
      } finally {
        this.deployFormLoading = false;
      }
    }
  }

  updatedSelectedApplication(applications: Application[], applicationId?: string | null) {
    this.selectedApplication = applications.find((a) => a.id === applicationId) || null;
  }

  openStatusModal(deploymentTarget: DeploymentTarget, modal: TemplateRef<any>) {
    const deployment = deploymentTarget.deployment;
    if (deployment?.id) {
      this.selectedDeploymentTarget.set(deploymentTarget);
      this.statuses = this.deploymentStatuses.pollStatuses(deployment);
      this.showModal(modal);
    }
  }

  async openInstructionsModal(deploymentTarget: DeploymentTarget, modal: TemplateRef<any>) {
    if (deploymentTarget.currentStatus !== undefined) {
      const message = `
        If you continue, the previous authentication secret for ${deploymentTarget.name} becomes invalid.
        Continue?`;
      const warning =
        deploymentTarget.createdBy?.userRole === 'customer' && this.auth.hasRole('vendor')
          ? {message: this.customerManagedWarning}
          : undefined;
      if (!(await firstValueFrom(this.overlay.confirm({message, warning})))) {
        return;
      }
    }
    this.showModal(modal);
  }

  public async updateDeploymentTargetAgent(dt: DeploymentTarget): Promise<void> {
    try {
      const agentVersions = await firstValueFrom(this.agentVersions$);
      if (agentVersions.length > 0) {
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

  private getAvailableVersions(deployment: DeploymentWithLatestRevision): Observable<ApplicationVersion[]> {
    return (
      deployment.applicationLicenseId
        ? this.licenses
            .list()
            .pipe(
              map((licenses) => licenses.find((license) => license.id === deployment.applicationLicenseId)?.versions)
            )
        : of(undefined)
    ).pipe(
      switchMap((versions) =>
        versions?.length
          ? of(versions)
          : this.applications$.pipe(
              map((apps) => apps.find((app) => app.id === deployment.applicationId)?.versions ?? [])
            )
      ),
      map((versions) => versions.filter((version) => !isArchived(version)))
    );
  }

  private findMaxVersion(versions: ApplicationVersion[]): ApplicationVersion | undefined {
    try {
      return maxBy(
        versions,
        (version) => new SemVer(version.name),
        (a, b) => a.compare(b) > 0
      );
    } catch (e) {
      console.warn('semver compare failed, falling back to creation date', e);
      return maxBy(versions, (version) => new Date(version.createdAt!));
    }
  }
}
