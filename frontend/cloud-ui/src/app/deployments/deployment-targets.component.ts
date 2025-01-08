import {GlobalPositionStrategy} from '@angular/cdk/overlay';
import {AsyncPipe, DatePipe, NgOptimizedImage} from '@angular/common';
import {
  AfterViewInit,
  Component,
  effect,
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
import {faMagnifyingGlass, faPen, faPlus, faShip, faTrash, faXmark} from '@fortawesome/free-solid-svg-icons';
import {
  catchError,
  combineLatest,
  filter,
  first,
  firstValueFrom,
  lastValueFrom,
  map,
  of,
  shareReplay,
  Subject,
  switchMap,
  takeUntil,
  tap,
  withLatestFrom,
} from 'rxjs';
import {RelativeDatePipe} from '../../util/dates';
import {filteredByFormControl} from '../../util/filter';
import {IsStalePipe} from '../../util/model';
import {drawerFlyInOut} from '../animations/drawer';
import {modalFlyInOut} from '../animations/modal';
import {ConnectInstructionsComponent} from '../components/connect-instructions/connect-instructions.component';
import {InstallationWizardComponent} from '../components/installation-wizard/installation-wizard.component';
import {StatusDotComponent} from '../components/status-dot';
import {ApplicationsService} from '../services/applications.service';
import {AuthService} from '../services/auth.service';
import {DeploymentTargetsService} from '../services/deployment-targets.service';
import {DeploymentService} from '../services/deployment.service';
import {DialogRef, OverlayService} from '../services/overlay.service';
import {ToastService} from '../services/toast.service';
import {Application} from '../types/application';
import {Deployment, DeploymentType} from '../types/deployment';
import {DeploymentTarget} from '../types/deployment-target';
import {DeploymentTargetViewModel} from './deployment-target-view-model';

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
    RelativeDatePipe,
    StatusDotComponent,
    ConnectInstructionsComponent,
    InstallationWizardComponent,
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

  readonly magnifyingGlassIcon = faMagnifyingGlass;
  readonly plusIcon = faPlus;
  readonly penIcon = faPen;
  readonly shipIcon = faShip;
  readonly xmarkIcon = faXmark;
  protected readonly faTrash = faTrash;

  private destroyed$ = new Subject();

  private modal?: DialogRef;
  private manageDeploymentTargetRef?: DialogRef;
  private deploymentWizardOverlayRef?: DialogRef;

  @ViewChild('deploymentWizard') wizardRef?: TemplateRef<unknown>;

  private selectedDeploymentTarget = signal<DeploymentTargetViewModel | null>(null);
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
  });

  readonly deployForm = new FormGroup({
    deploymentTargetId: new FormControl<string | undefined>(undefined, Validators.required),
    applicationId: new FormControl<string | undefined>(undefined, Validators.required),
    applicationVersionId: new FormControl<string | undefined>({value: undefined, disabled: true}, Validators.required),
    valuesYaml: new FormControl<string | undefined>({value: undefined, disabled: true}),
    notes: new FormControl<string | undefined>(undefined),
  });

  readonly deploymentTargets$ = this.deploymentTargets.list().pipe(
    shareReplay(1),
    map((dts) =>
      dts.map((dt) => {
        let dtView = dt as DeploymentTargetViewModel;
        if (dtView.id) {
          dtView.latestDeployment = this.deploymentTargets.latestDeploymentFor(dtView.id);
        }
        return dtView;
      })
    )
  );
  readonly filteredDeploymentTargets$ = filteredByFormControl(
    this.deploymentTargets$,
    this.filterForm.controls.search,
    (dt, search) => !search || (dt.name || '').toLowerCase().includes(search.toLowerCase())
  );

  private readonly applications$ = this.applications.list();
  readonly filteredApplications$ = combineLatest([
    this.applications$,
    toObservable(this.selectedDeploymentTarget),
  ]).pipe(map(([apps, dt]) => apps.filter((app) => app.type === dt?.type)));

  ngOnInit() {
    this.registerDeployFormChanges();
  }

  ngAfterViewInit() {
    if (this.fullVersion) {
      combineLatest([this.filteredApplications$, this.deploymentTargets$])
        .pipe(first())
        .subscribe(([apps, dts]) => {
          if (this.auth.hasRole('customer') && apps.length > 0 && dts.length === 0) {
            this.openWizard();
          }
        });
    }
  }

  ngOnDestroy(): void {
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
        switchMap(() => this.deploymentTargets.delete(dt))
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
    if (this.editForm.valid) {
      const val = this.editForm.value;
      const dt: DeploymentTarget = {
        id: val.id!,
        name: val.name!,
        type: val.type!,
      };

      if (val.geolocation?.lat && val.geolocation.lon) {
        dt.geolocation = {
          lat: val.geolocation.lat,
          lon: val.geolocation.lon,
        };
      }

      this.loadDeploymentTarget(
        await lastValueFrom(val.id ? this.deploymentTargets.update(dt) : this.deploymentTargets.create(dt))
      );
      this.toast.success(`${dt.name} saved successfully`);
      this.hideDrawer();
    } else {
      this.editForm.markAllAsTouched();
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
  }

  async newDeployment(deploymentTarget: DeploymentTargetViewModel, modalTemplate: TemplateRef<any>) {
    const apps = await firstValueFrom(this.applications$);
    const latestDeployment = await firstValueFrom(
      deploymentTarget.latestDeployment?.pipe(catchError(() => of(null))) ?? of(null)
    );
    this.selectedDeploymentTarget.set(deploymentTarget);
    this.deployForm.reset({
      deploymentTargetId: deploymentTarget.id,
      applicationId: latestDeployment?.applicationId,
      applicationVersionId: latestDeployment?.applicationVersionId,
    });
    if (latestDeployment) {
      this.updatedSelectedApplication(apps, latestDeployment.applicationId);
    }
    this.showModal(modalTemplate);
  }

  async saveDeployment() {
    if (this.deployForm.valid) {
      const deployment = this.deployForm.value;
      await firstValueFrom(this.deployments.create(deployment as Deployment));
      this.selectedDeploymentTarget()!.latestDeployment = this.deploymentTargets.latestDeploymentFor(
        this.selectedDeploymentTarget()!.id!
      );
      this.toast.success('Deployment saved successfully');
      this.hideModal();
    } else {
      this.deployForm.markAllAsTouched();
    }
  }

  updatedSelectedApplication(applications: Application[], applicationId?: string | null) {
    this.selectedApplication = applications.find((a) => a.id === applicationId) || null;
  }
}
