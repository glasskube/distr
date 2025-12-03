import {CdkConnectedOverlay, CdkOverlayOrigin} from '@angular/cdk/overlay';
import {AsyncPipe, NgTemplateOutlet} from '@angular/common';
import {Component, inject, signal, WritableSignal} from '@angular/core';
import {toSignal} from '@angular/core/rxjs-interop';
import {RouterLink, RouterLinkActive} from '@angular/router';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {
  faAddressBook,
  faArrowRightLong,
  faBox,
  faBoxesStacked,
  faChevronDown,
  faCreditCard,
  faDashboard,
  faGear,
  faHome,
  faKey,
  faLightbulb,
  faPalette,
  faUsers,
} from '@fortawesome/free-solid-svg-icons';
import {map} from 'rxjs';
import {buildConfig} from '../../../buildconfig';
import {RequireCustomerDirective, RequireVendorDirective} from '../../directives/required-role.directive';
import {AuthService} from '../../services/auth.service';
import {FeatureFlagService} from '../../services/feature-flag.service';
import {OrganizationService} from '../../services/organization.service';
import {SidebarService} from '../../services/sidebar.service';
import {TutorialsService} from '../../services/tutorials.service';

@Component({
  selector: 'app-side-bar',
  standalone: true,
  templateUrl: './side-bar.component.html',
  imports: [
    RouterLink,
    FaIconComponent,
    AsyncPipe,
    RouterLinkActive,
    CdkOverlayOrigin,
    CdkConnectedOverlay,
    NgTemplateOutlet,
    RequireVendorDirective,
    RequireCustomerDirective,
  ],
})
export class SideBarComponent {
  protected readonly auth = inject(AuthService);
  protected readonly sidebar = inject(SidebarService);
  protected readonly featureFlags = inject(FeatureFlagService);
  protected readonly tutorialsService = inject(TutorialsService);
  private readonly organizationService = inject(OrganizationService);

  protected readonly buildConfig = buildConfig;

  protected readonly faDashboard = faDashboard;
  protected readonly faBoxesStacked = faBoxesStacked;
  protected readonly faLightbulb = faLightbulb;
  protected readonly faKey = faKey;
  protected readonly faGear = faGear;
  protected readonly faUsers = faUsers;
  protected readonly faPalette = faPalette;
  protected readonly faAddressBook = faAddressBook;
  protected readonly faBox = faBox;
  protected readonly faCreditCard = faCreditCard;
  protected readonly faArrowRightLong = faArrowRightLong;
  protected readonly faHome = faHome;
  protected readonly faChevronDown = faChevronDown;

  protected feedbackAlert = true;
  protected readonly agentsSubMenuOpen = signal(true);
  protected readonly licenseSubMenuOpen = signal(true);
  protected readonly registrySubMenuOpen = signal(true);
  protected readonly licenseOverlayOpen = signal(false);

  protected readonly organization$ = this.organizationService.get();

  protected readonly isTrial = toSignal(this.organization$.pipe(map((org) => org.subscriptionType === 'trial')));

  protected toggle(signal: WritableSignal<boolean>) {
    signal.update((val) => !val);
  }
}
