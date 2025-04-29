import {OverlayModule} from '@angular/cdk/overlay';
import {Component, inject, input} from '@angular/core';
import {ReactiveFormsModule} from '@angular/forms';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {
  faBox,
  faCircleExclamation,
  faHeartPulse,
  faPen,
  faShip,
  faTrash,
  faUserCircle,
  faXmark,
} from '@fortawesome/free-solid-svg-icons';
import {drawerFlyInOut} from '../../animations/drawer';
import {dropdownAnimation} from '../../animations/dropdown';
import {modalFlyInOut} from '../../animations/modal';
import {AuthService} from '../../services/auth.service';
import {UserAccountWithRole} from '@glasskube/distr-sdk';
import {ArtifactWithTags} from '../../services/artifacts.service';

@Component({
  selector: 'app-artifacts-by-customer-card',
  templateUrl: './artifacts-by-customer-card.component.html',
  imports: [FaIconComponent, OverlayModule, ReactiveFormsModule],
  animations: [modalFlyInOut, drawerFlyInOut, dropdownAnimation],
})
export class ArtifactsByCustomerCardComponent {
  protected readonly auth = inject(AuthService);

  public readonly customer = input.required<UserAccountWithRole>();
  public readonly artifacts = input.required<ArtifactWithTags[]>();

  protected readonly faShip = faShip;
  protected readonly faPen = faPen;
  protected readonly faTrash = faTrash;
  protected readonly faHeartPulse = faHeartPulse;
  protected readonly faXmark = faXmark;
  protected readonly faCircleExclamation = faCircleExclamation;
  protected readonly faUserCircle = faUserCircle;
  protected readonly faBox = faBox;
}
