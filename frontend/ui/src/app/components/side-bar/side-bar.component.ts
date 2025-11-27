import {Component, effect, ElementRef, inject, signal, ViewChild, WritableSignal} from '@angular/core';
import {RouterLink, RouterLinkActive} from '@angular/router';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {
  faAddressBook,
  faArrowRightLong,
  faBox,
  faBoxesStacked,
  faCheckDouble,
  faChevronDown,
  faCodeFork,
  faCreditCard,
  faDashboard,
  faGear,
  faHome,
  faKey,
  faLightbulb,
  faPalette,
  faServer,
  faUserCheck,
  faUsers,
} from '@fortawesome/free-solid-svg-icons';
import {RequireRoleDirective} from '../../directives/required-role.directive';
import {SidebarService} from '../../services/sidebar.service';
import {buildConfig} from '../../../buildconfig';
import {FeatureFlagService} from '../../services/feature-flag.service';
import {AsyncPipe, NgTemplateOutlet} from '@angular/common';
import {CdkConnectedOverlay, CdkOverlayOrigin} from '@angular/cdk/overlay';
import {TutorialsService} from '../../services/tutorials.service';
import {OrganizationService} from '../../services/organization.service';

@Component({
  selector: 'app-side-bar',
  standalone: true,
  templateUrl: './side-bar.component.html',
  imports: [
    RouterLink,
    FaIconComponent,
    RequireRoleDirective,
    AsyncPipe,
    RouterLinkActive,
    CdkOverlayOrigin,
    CdkConnectedOverlay,
    NgTemplateOutlet,
  ],
})
export class SideBarComponent {
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

  protected toggle(signal: WritableSignal<boolean>) {
    signal.update((val) => !val);
  }
}
