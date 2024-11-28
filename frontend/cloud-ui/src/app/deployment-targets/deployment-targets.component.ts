import {AsyncPipe, DatePipe, NgOptimizedImage} from '@angular/common';
import {Component, inject, Input, TemplateRef, ViewContainerRef} from '@angular/core';
import {FormControl, FormGroup, FormsModule, ReactiveFormsModule, Validators} from '@angular/forms';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faCaretDown, faMagnifyingGlass, faPen, faPlus, faTrash, faXmark} from '@fortawesome/free-solid-svg-icons';
import {lastValueFrom} from 'rxjs';
import {RelativeDatePipe} from '../../util/dates';
import {IsStalePipe} from '../../util/model';
import {modalFlyInOut} from '../animations/modal';
import {DeploymentTargetsService} from '../services/deployment-targets.service';
import {EmbeddedOverlayRef, OverlayService} from '../services/overlay.service';
import {DeploymentTarget} from '../types/deployment-target';
import {StatusDotComponent} from '../components/status-dot';
import {drawerFlyInOut} from '../animations/drawer';

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
  ],
  templateUrl: './deployment-targets.component.html',
  animations: [modalFlyInOut, drawerFlyInOut],
})
export class DeploymentTargetsComponent {
  @Input('fullVersion') fullVersion = false;
  readonly magnifyingGlassIcon = faMagnifyingGlass;
  readonly plusIcon = faPlus;
  readonly caretDownIcon = faCaretDown;
  readonly penIcon = faPen;
  readonly trashIcon = faTrash;
  readonly xmarkIcon = faXmark;

  private instructionsModal?: EmbeddedOverlayRef;
  private manageDeploymentTargetRef?: EmbeddedOverlayRef;
  private readonly overlay = inject(OverlayService);
  private readonly viewContainerRef = inject(ViewContainerRef);

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

  openDrawer(templateRef: TemplateRef<unknown>, deploymentTarget?: DeploymentTarget) {
    this.hideDrawer();
    if (deploymentTarget) {
      this.loadDeploymentTarget(deploymentTarget);
    } else {
      this.reset();
    }
    this.manageDeploymentTargetRef = this.overlay.showDrawer(templateRef, this.viewContainerRef);
  }

  hideDrawer() {
    this.manageDeploymentTargetRef?.close();
    this.reset();
  }

  reset() {
    this.editForm.reset();
    this.editForm.patchValue({type: 'docker'});
  }

  loadDeploymentTarget(dt: DeploymentTarget) {
    this.editForm.patchValue({
      // to reset the geolocation inputs in case dt has no geolocation
      geolocation: {lat: undefined, lon: undefined},
      ...dt,
    });
  }

  showInstructions(templateRef: TemplateRef<unknown>) {
    this.hideInstructions();
    this.instructionsModal = this.overlay.showModal(templateRef, this.viewContainerRef);
  }

  hideInstructions(): void {
    this.instructionsModal?.close();
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
    }
  }
}
