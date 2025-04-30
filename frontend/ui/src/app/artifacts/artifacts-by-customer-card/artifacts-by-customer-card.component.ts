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
import {ApplicationVersion, UserAccountWithRole} from '@glasskube/distr-sdk';
import {DashboardArtifact} from '../../services/dashboard.service';
import {maxBy} from '../../../util/arrays';
import {SemVer} from 'semver';
import {TaggedArtifactVersion} from '../../services/artifacts.service';
import {AsyncPipe} from '@angular/common';
import {SecureImagePipe} from '../../../util/secureImage';

@Component({
  selector: 'app-artifacts-by-customer-card',
  templateUrl: './artifacts-by-customer-card.component.html',
  imports: [FaIconComponent, OverlayModule, ReactiveFormsModule, AsyncPipe, SecureImagePipe],
  animations: [modalFlyInOut, drawerFlyInOut, dropdownAnimation],
})
export class ArtifactsByCustomerCardComponent {
  protected readonly auth = inject(AuthService);

  public readonly customer = input.required<UserAccountWithRole>();
  public readonly artifacts = input.required<DashboardArtifact[]>();

  protected isOnLatest(artifact: DashboardArtifact): boolean {
    const max = this.findMaxVersion(artifact.artifact.versions ?? []);
    const includesPulled = max?.tags.map((t) => t.name).includes(artifact.latestPulledVersion) ?? false;
    if (includesPulled) {
      return true;
    }
    const pulledStillAvailable = (artifact.artifact.versions ?? []).find((v) =>
      v.tags.find((t) => t.name === artifact.latestPulledVersion)
    );
    return !pulledStillAvailable;
  }

  private findMaxVersion(versions: TaggedArtifactVersion[]): TaggedArtifactVersion | undefined {
    try {
      return maxBy(
        versions,
        (version) => this.getFirstSemverTag(version),
        (a, b) => a.compare(b) > 0
      );
    } catch (e) {
      return maxBy(versions, (version) => new Date(version.createdAt!));
    }
  }

  private getFirstSemverTag(version: TaggedArtifactVersion): SemVer {
    for (let v of version.tags) {
      try {
        return new SemVer(v.name);
      } catch (e) {}
    }
    throw 'no semver tag found';
  }

  protected readonly faShip = faShip;
  protected readonly faPen = faPen;
  protected readonly faTrash = faTrash;
  protected readonly faHeartPulse = faHeartPulse;
  protected readonly faXmark = faXmark;
  protected readonly faCircleExclamation = faCircleExclamation;
  protected readonly faUserCircle = faUserCircle;
  protected readonly faBox = faBox;
}
