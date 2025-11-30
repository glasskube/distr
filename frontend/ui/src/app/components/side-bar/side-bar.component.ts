import {Component, effect, ElementRef, inject, signal, viewChild, ViewChild, WritableSignal} from '@angular/core';
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
import {
  RequireCustomerDirective,
  RequireRoleDirective,
  RequireVendorDirective,
} from '../../directives/required-role.directive';
import {SidebarService} from '../../services/sidebar.service';
import {buildConfig} from '../../../buildconfig';
import {FeatureFlagService} from '../../services/feature-flag.service';
import {AsyncPipe, NgTemplateOutlet} from '@angular/common';
import {CdkConnectedOverlay, CdkOverlayOrigin} from '@angular/cdk/overlay';
import {TutorialsService} from '../../services/tutorials.service';
import {AuthService} from '../../services/auth.service';

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
  protected readonly buildConfig = buildConfig;
  protected readonly faArrowRightLong = faArrowRightLong;
  protected readonly faHome = faHome;
  protected readonly faChevronDown = faChevronDown;

  protected feedbackAlert = true;
  protected readonly agentsSubMenuOpen = signal(true);
  protected readonly licenseSubMenuOpen = signal(true);
  protected readonly registrySubMenuOpen = signal(true);
  protected readonly licenseOverlayOpen = signal(false);

  private readonly asideElement = viewChild.required<ElementRef<HTMLElement>>('asideElement');

  constructor() {
    effect(() => {
      const show = this.sidebar.showSidebar();
      this.asideElement().nativeElement.classList.toggle('translate-x-0', show);
      this.asideElement().nativeElement.classList.toggle('-translate-x-full', !show);
    });
  }

  protected toggle(signal: WritableSignal<boolean>) {
    signal.update((val) => !val);
  }

  protected readonly faUserCheck = faUserCheck;
}
