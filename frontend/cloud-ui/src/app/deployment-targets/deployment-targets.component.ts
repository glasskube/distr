import {Component, inject, Input} from '@angular/core';
import {DeploymentTargetsService} from '../services/deployment-targets.service';
import {AsyncPipe, DatePipe, NgOptimizedImage} from '@angular/common';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {FormControl, FormGroup, FormsModule, ReactiveFormsModule, Validators} from '@angular/forms';
import {faCaretDown, faMagnifyingGlass, faPen, faPlus, faTrash, faXmark} from '@fortawesome/free-solid-svg-icons';
import {DeploymentTarget} from '../types/deployment-target';
import {lastValueFrom} from 'rxjs';

@Component({
  selector: 'app-deployment-targets',
  imports: [AsyncPipe, DatePipe, FaIconComponent, FormsModule, ReactiveFormsModule, NgOptimizedImage],
  templateUrl: './deployment-targets.component.html',
})
export class DeploymentTargetsComponent {
  @Input('fullVersion') fullVersion = false;
  readonly magnifyingGlassIcon = faMagnifyingGlass;
  readonly plusIcon = faPlus;
  readonly caretDownIcon = faCaretDown;
  readonly penIcon = faPen;
  readonly trashIcon = faTrash;
  readonly xmarkIcon = faXmark;

  private readonly deploymentTargets = inject(DeploymentTargetsService);
  readonly deploymentTargets$ = this.deploymentTargets.list();

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
}
