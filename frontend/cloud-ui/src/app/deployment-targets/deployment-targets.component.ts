import {Component, inject} from '@angular/core';
import {DeploymentTargetsService} from '../services/deployment-targets.service';
import {AsyncPipe, DatePipe, JsonPipe, NgOptimizedImage} from '@angular/common';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {FormControl, FormGroup, FormsModule, ReactiveFormsModule, Validators} from '@angular/forms';
import {
  faCaretDown,
  faMagnifyingGlass,
  faPen,
  faPlus,
  faShip,
  faTrash,
  faXmark
} from '@fortawesome/free-solid-svg-icons';
import {DeploymentTarget} from '../types/deployment-target';
import {firstValueFrom, lastValueFrom, map, of} from 'rxjs';
import {ApplicationsService} from '../applications/applications.service';
import {Application} from '../types/application';
import {DeploymentService} from '../services/deployment.service';
import {Deployment} from '../types/deployment';

@Component({
  selector: 'app-deployment-targets',
  imports: [AsyncPipe, DatePipe, FaIconComponent, FormsModule, ReactiveFormsModule, NgOptimizedImage],
  templateUrl: './deployment-targets.component.html',
  standalone: true,
})
export class DeploymentTargetsComponent {
  readonly magnifyingGlassIcon = faMagnifyingGlass;
  readonly plusIcon = faPlus;
  readonly caretDownIcon = faCaretDown;
  readonly penIcon = faPen;
  readonly shipIcon = faShip;
  readonly trashIcon = faTrash;
  readonly xmarkIcon = faXmark;

  private readonly deploymentTargets = inject(DeploymentTargetsService);
  readonly deploymentTargets$ = this.deploymentTargets.list()
    .pipe(map(dts =>
      dts.map(dt => {
        if (dt.id) {
          dt.latestDeployment = this.deploymentTargets.latestDeploymentFor(dt.id);
        }
        return dt;
      })));

  editForm = new FormGroup({
    id: new FormControl<string | undefined>(undefined),
    name: new FormControl('', Validators.required),
    type: new FormControl('', Validators.required),
    geolocation: new FormGroup({
      lat: new FormControl<number | undefined>(undefined),
      lon: new FormControl<number | undefined>(undefined),
    }),
  });

  newDeploymentTarget() {
    this.editForm.reset();
  }

  editDeploymentTarget(dt: DeploymentTarget) {
    this.editForm.patchValue({
      // to reset the geolocation inputs in case dt has no geolocation
      geolocation: {lat: undefined, lon: undefined},
      ...dt,
    });
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

      this.editDeploymentTarget(
        await lastValueFrom(val.id ? this.deploymentTargets.update(dt) : this.deploymentTargets.create(dt))
      );
    }
  }

  private selectedDeploymentTarget?: DeploymentTarget | null;
  private readonly applications = inject(ApplicationsService);
  readonly applications$ = this.applications.list();
  selectedApplication?: Application | null;

  readonly deployments = inject(DeploymentService);


  deployForm = new FormGroup({
    deploymentTargetId: new FormControl<string | undefined>(undefined, Validators.required),
    applicationId: new FormControl<string | undefined>(undefined, Validators.required),
    applicationVersionId: new FormControl<string | undefined>(undefined, Validators.required),
    notes: new FormControl<string | undefined>(undefined)
  });

  isDeploymentTargetHealthy(dt: DeploymentTarget) {
    // TODO
    return true;
  }

  async newDeployment(dt: DeploymentTarget) {
    this.deployForm.reset();
    this.selectedDeploymentTarget = dt;
    this.deployForm.patchValue({
      deploymentTargetId: dt.id,
    });
    this.deploymentTargets.latestDeploymentFor(dt.id!!).subscribe(d => {
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
      this.selectedDeploymentTarget!!.latestDeployment =
        this.deploymentTargets.latestDeploymentFor(this.selectedDeploymentTarget!!.id!!);
    }
  }

  onApplicationIdChange($event: any) {
    this.updatedSelectedApplication($event.target.value);
  }

  async updatedSelectedApplication(applicationId: string) {
    let applications = await firstValueFrom(this.applications$);
    this.selectedApplication = applications.find(a => a.id === applicationId) || null;
  }

  validApplicationSelected() {
    return Boolean(this.deployForm.value.applicationId);
  }
}
