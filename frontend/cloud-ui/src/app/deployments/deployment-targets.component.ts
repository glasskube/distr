import {Component, inject, Input, OnDestroy, OnInit, TemplateRef, ViewChild, ViewContainerRef} from '@angular/core';
import {DeploymentTargetsService} from '../services/deployment-targets.service';
import {AsyncPipe, DatePipe, NgOptimizedImage} from '@angular/common';
import {FormControl, FormGroup, FormsModule, ReactiveFormsModule, Validators} from '@angular/forms';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faMagnifyingGlass, faPen, faPlus, faShip, faTrash, faXmark} from '@fortawesome/free-solid-svg-icons';
import {combineLatest, first, firstValueFrom, lastValueFrom, map} from 'rxjs';
import {RelativeDatePipe} from '../../util/dates';
import {IsStalePipe} from '../../util/model';
import {modalFlyInOut} from '../animations/modal';
import {EmbeddedOverlayRef, OverlayService} from '../services/overlay.service';
import {DeploymentTarget} from '../types/deployment-target';
import {Application} from '../types/application';
import {DeploymentService} from '../services/deployment.service';
import {Deployment} from '../types/deployment';
import {StatusDotComponent} from '../components/status-dot';
import {drawerFlyInOut} from '../animations/drawer';
import {ApplicationsService} from '../services/applications.service';
import {DeploymentTargetViewModel} from './DeploymentTargetViewModel';
import {ConnectInstructionsComponent} from '../components/connect-instructions/connect-instructions.component';
import {InstallationWizardComponent} from '../components/installation-wizard/installation-wizard.component';
import {GlobalPositionStrategy} from '@angular/cdk/overlay';
import {AuthService} from '../services/auth.service';
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
    RelativeDatePipe,
    StatusDotComponent,
    ConnectInstructionsComponent,
    InstallationWizardComponent,
  ],
  templateUrl: './deployment-targets.component.html',
  standalone: true,
  animations: [modalFlyInOut, drawerFlyInOut],
})
export class DeploymentTargetsComponent implements OnInit, OnDestroy {
  private readonly toast = inject(ToastService);
  private readonly auth = inject(AuthService);
  @Input('fullVersion') fullVersion = false;
  readonly magnifyingGlassIcon = faMagnifyingGlass;
  readonly plusIcon = faPlus;
  readonly penIcon = faPen;
  readonly shipIcon = faShip;
  readonly xmarkIcon = faXmark;
  protected readonly faTrash = faTrash;

  private modal?: EmbeddedOverlayRef;
  private manageDeploymentTargetRef?: EmbeddedOverlayRef;
  private readonly overlay = inject(OverlayService);
  private readonly viewContainerRef = inject(ViewContainerRef);

  @ViewChild('deploymentWizard') wizardRef?: TemplateRef<unknown>;
  private deploymentWizardOverlayRef?: EmbeddedOverlayRef;

  private readonly deploymentTargets = inject(DeploymentTargetsService);
  readonly deploymentTargets$ = this.deploymentTargets.list().pipe(
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

  editForm = new FormGroup({
    id: new FormControl<string | undefined>(undefined),
    name: new FormControl('', Validators.required),
    type: new FormControl('', Validators.required),
    geolocation: new FormGroup({
      lat: new FormControl<number | undefined>(undefined),
      lon: new FormControl<number | undefined>(undefined),
    }),
  });

  ngOnInit() {
    if (this.fullVersion) {
      const always = false;
      const isCustomer = this.auth.getClaims().role === 'customer';
      combineLatest([this.applications$, this.deploymentTargets$])
        .pipe(first())
        .subscribe(([apps, dts]) => {
          if (always || (isCustomer && apps.length > 0 && dts.length === 0)) {
            this.openWizard();
          }
        });
    }
  }

  openDrawer(templateRef: TemplateRef<unknown>, deploymentTarget: DeploymentTarget) {
    this.hideDrawer();
    if (deploymentTarget) {
      this.loadDeploymentTarget(deploymentTarget);
    }
    this.manageDeploymentTargetRef = this.overlay.showDrawer(templateRef, this.viewContainerRef);
  }

  hideDrawer() {
    this.manageDeploymentTargetRef?.close();
    this.reset();
  }

  openWizard() {
    this.deploymentWizardOverlayRef?.close();
    this.deploymentWizardOverlayRef = this.overlay.showModal(this.wizardRef!, this.viewContainerRef, {
      hasBackdrop: true,
      backdropStyleOnly: true,
      positionStrategy: new GlobalPositionStrategy().centerHorizontally().centerVertically(),
    });
  }

  closeWizard() {
    this.deploymentWizardOverlayRef?.close();
  }

  reset() {
    this.editForm.reset();
    this.editForm.patchValue({type: 'docker'});
  }

  async deleteDeploymentTarget(dt: DeploymentTarget) {
    if (confirm(`Really delete ${dt.name}? This action can not be undone.`)) {
      await firstValueFrom(this.deploymentTargets.delete(dt));
    }
  }

  loadDeploymentTarget(dt: DeploymentTarget) {
    this.editForm.patchValue({
      // to reset the geolocation inputs in case dt has no geolocation
      geolocation: {lat: undefined, lon: undefined},
      ...dt,
    });
  }

  showDeploymentModal(templateRef: TemplateRef<unknown>) {
    this.hideModal();
    this.modal = this.overlay.showModal(templateRef, this.viewContainerRef);
  }

  async showInstructionsModal(templateRef: TemplateRef<unknown>, dt: DeploymentTarget) {
    this.hideModal();
    this.modal = this.overlay.showModal(templateRef, this.viewContainerRef);
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

  private selectedDeploymentTarget?: DeploymentTargetViewModel | null;
  private readonly applications = inject(ApplicationsService);
  readonly applications$ = this.applications.list();
  selectedApplication?: Application | null;

  deployForm = new FormGroup({
    deploymentTargetId: new FormControl<string | undefined>(undefined, Validators.required),
    applicationId: new FormControl<string | undefined>(undefined, Validators.required),
    applicationVersionId: new FormControl<string | undefined>(undefined, Validators.required),
    notes: new FormControl<string | undefined>(undefined),
  });

  readonly deployments = inject(DeploymentService);

  private readonly applicationIdChange$ = this.deployForm.controls.applicationId.valueChanges.subscribe((it) =>
    this.updatedSelectedApplication(it!!)
  );

  ngOnDestroy(): void {
    this.applicationIdChange$.unsubscribe();
  }

  async newDeployment(dt: DeploymentTargetViewModel, deploymentModal: TemplateRef<any>) {
    this.showDeploymentModal(deploymentModal);
    this.deployForm.reset();
    this.selectedDeploymentTarget = dt;
    this.deployForm.patchValue({
      deploymentTargetId: dt.id,
    });
    this.deploymentTargets.latestDeploymentFor(dt.id!!).subscribe((d) => {
      this.deployForm.patchValue({
        applicationId: d.applicationId,
        applicationVersionId: d.applicationVersionId,
      });
      this.updatedSelectedApplication(d.applicationId);
    });
  }

  async saveDeployment() {
    if (this.deployForm.valid) {
      const deployment = this.deployForm.value;
      await firstValueFrom(this.deployments.create(deployment as Deployment));
      this.selectedDeploymentTarget!!.latestDeployment = this.deploymentTargets.latestDeploymentFor(
        this.selectedDeploymentTarget!!.id!!
      );
      this.toast.success('Deployment saved successfully');
      this.hideModal();
    } else {
      this.deployForm.markAllAsTouched();
    }
  }

  async updatedSelectedApplication(applicationId: string) {
    let applications = await firstValueFrom(this.applications$);
    this.selectedApplication = applications.find((a) => a.id === applicationId) || null;
  }
}
