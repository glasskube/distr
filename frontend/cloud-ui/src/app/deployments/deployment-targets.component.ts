import {GlobalPositionStrategy} from '@angular/cdk/overlay';
import {AsyncPipe, DatePipe, NgOptimizedImage, UpperCasePipe} from '@angular/common';
import {
  AfterViewInit,
  Component,
  inject,
  Input,
  OnDestroy,
  OnInit,
  signal,
  TemplateRef,
  ViewChild,
} from '@angular/core';
import {toObservable} from '@angular/core/rxjs-interop';
import {FormControl, FormGroup, FormsModule, ReactiveFormsModule, Validators} from '@angular/forms';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {
  faHeartPulse,
  faMagnifyingGlass,
  faPen,
  faPlus,
  faShip,
  faTrash,
  faXmark,
} from '@fortawesome/free-solid-svg-icons';
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
  takeUntil,
  tap,
  withLatestFrom,
} from 'rxjs';
import {getFormDisplayedError} from '../../util/errors';
import {filteredByFormControl} from '../../util/filter';
import {IsStalePipe} from '../../util/model';
import {drawerFlyInOut} from '../animations/drawer';
import {modalFlyInOut} from '../animations/modal';
import {ConnectInstructionsComponent} from '../components/connect-instructions/connect-instructions.component';
import {InstallationWizardComponent} from '../components/installation-wizard/installation-wizard.component';
import {StatusDotComponent} from '../components/status-dot';
import {YamlEditorComponent} from '../components/yaml-editor.component';
import {AgentVersionService} from '../services/agent-version.service';
import {ApplicationsService} from '../services/applications.service';
import {AuthService} from '../services/auth.service';
import {DeploymentTargetsService} from '../services/deployment-targets.service';
import {DeploymentService} from '../services/deployment.service';
import {DialogRef, OverlayService} from '../services/overlay.service';
import {ToastService} from '../services/toast.service';
import {Application} from '../types/application';
import {DeploymentRequest, DeploymentRevisionStatus, DeploymentTargetScope, DeploymentType} from '../types/deployment';
import {DeploymentTarget} from '../types/deployment-target';

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
    YamlEditorComponent,
  ],
  templateUrl: './deployment-targets.component.html',
  standalone: true,
  animations: [modalFlyInOut, drawerFlyInOut],
})
export class DeploymentTargetsComponent implements OnInit, AfterViewInit, OnDestroy {
  @Input('fullVersion') fullVersion = false;

  public readonly auth = inject(AuthService);
  private readonly toast = inject(ToastService);
  private readonly overlay = inject(OverlayService);
  private readonly applications = inject(ApplicationsService);
  private readonly deploymentTargets = inject(DeploymentTargetsService);
  private readonly deployments = inject(DeploymentService);
  private readonly agentVersions = inject(AgentVersionService);

  readonly magnifyingGlassIcon = faMagnifyingGlass;
  readonly plusIcon = faPlus;
  readonly penIcon = faPen;
  readonly shipIcon = faShip;
  readonly xmarkIcon = faXmark;
  protected readonly faHeartPulse = faHeartPulse;
  protected readonly faTrash = faTrash;

  private destroyed$ = new Subject<void>();
  private modal?: DialogRef;
  private manageDeploymentTargetRef?: DialogRef;

  private deploymentWizardOverlayRef?: DialogRef;

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
  readonly deployForm = new FormGroup({
    deploymentTargetId: new FormControl<string | undefined>(undefined, Validators.required),
    deploymentId: new FormControl<string | undefined>(undefined),
    applicationId: new FormControl<string | undefined>(undefined, Validators.required),
    applicationVersionId: new FormControl<string | undefined>({value: undefined, disabled: true}, Validators.required),
    valuesYaml: new FormControl<string | undefined>({value: undefined, disabled: true}),
    releaseName: new FormControl<string>({value: '', disabled: true}, Validators.required),
    notes: new FormControl<string | undefined>(undefined),
  });

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

  ngOnInit() {
    this.registerDeployFormChanges();
  }

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
    if (deploymentTarget) {
      this.loadDeploymentTarget(deploymentTarget);
    }
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
    this.overlay
      .confirm(`Really delete ${dt.name}? This action can not be undone.`)
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

  loadDeploymentTarget(dt: DeploymentTarget) {
    this.editForm.patchValue({
      // to reset the geolocation inputs in case dt has no geolocation
      geolocation: {lat: undefined, lon: undefined},
      ...dt,
    });
  }

  showModal(templateRef: TemplateRef<unknown>) {
    this.hideModal();
    this.modal = this.overlay.showModal(templateRef);
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

  private registerDeployFormChanges() {
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
  }

  async newDeployment(deploymentTarget: DeploymentTarget, modalTemplate: TemplateRef<any>) {
    const apps = await firstValueFrom(this.applications$);
    this.selectedDeploymentTarget.set(deploymentTarget);
    this.deployForm.reset({
      deploymentTargetId: deploymentTarget.id,
      deploymentId: deploymentTarget.deployment?.id,
      applicationId: deploymentTarget.deployment?.applicationId,
      applicationVersionId: deploymentTarget.deployment?.applicationVersionId,
      releaseName: deploymentTarget.deployment?.releaseName,
    });
    if (deploymentTarget.deployment) {
      this.updatedSelectedApplication(apps, deploymentTarget.deployment.applicationId);
    }
    this.showModal(modalTemplate);
  }

  async saveDeployment() {
    this.deployForm.markAllAsTouched();
    if (this.deployForm.valid) {
      this.deployFormLoading = true;
      const deployment = this.deployForm.value;
      if (deployment.valuesYaml) {
        deployment.valuesYaml = btoa(deployment.valuesYaml);
      }
      try {
        await firstValueFrom(this.deployments.createOrUpdate(deployment as DeploymentRequest));
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
    if (!(this.auth.hasRole('vendor') && deploymentTarget.createdBy?.userRole === 'customer')) {
      const deployment = deploymentTarget.deployment;
      if (deployment?.id) {
        this.selectedDeploymentTarget.set(deploymentTarget);
        this.statuses = this.deployments.pollStatuses(deployment);
        this.showModal(modal);
      }
    }
  }

  async openInstructionsModal(deploymentTarget: DeploymentTarget, modal: TemplateRef<any>) {
    if (deploymentTarget.currentStatus !== undefined) {
      if (
        !(await firstValueFrom(
          this.overlay.confirm(
            `Warning: If you continue, the previous authentication secret for ${deploymentTarget.name} becomes invalid. Continue?`
          )
        ))
      ) {
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
}
