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
  public readonly sidebar = inject(SidebarService);
  public readonly featureFlags = inject(FeatureFlagService);
  protected readonly tutorialsService = inject(TutorialsService);
  public feedbackAlert = true;
  protected readonly faDashboard = faDashboard;
  protected readonly faBoxesStacked = faBoxesStacked;
  protected readonly faServer = faServer;
  protected readonly faLightbulb = faLightbulb;
  protected readonly faKey = faKey;
  protected readonly faGear = faGear;
  protected readonly faUsers = faUsers;
  protected readonly faCheckDouble = faCheckDouble;
  protected readonly faPalette = faPalette;
  protected readonly faAddressBook = faAddressBook;
  protected readonly faBox = faBox;
  protected readonly faCreditCard = faCreditCard;
  protected readonly buildConfig = buildConfig;
  protected readonly faArrowRightLong = faArrowRightLong;
  protected readonly faHome = faHome;
  protected readonly faChevronDown = faChevronDown;

  @ViewChild('asideElement') private asideElement?: ElementRef<HTMLElement>;
  protected readonly agentsSubMenuOpen = signal(true);
  protected readonly licenseSubMenuOpen = signal(true);
  protected readonly registrySubMenuOpen = signal(true);
  protected readonly licenseOverlayOpen = signal(false);

  constructor() {
    effect(() => {
      const show = this.sidebar.showSidebar();
      this.asideElement?.nativeElement.classList.toggle('translate-x-0', show);
      this.asideElement?.nativeElement.classList.toggle('-translate-x-full', !show);
    });
  }

  protected toggle(signal: WritableSignal<boolean>) {
    signal.update((val) => !val);
  }

  protected readonly faUserCheck = faUserCheck;
}
