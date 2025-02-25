import {Component, effect, ElementRef, inject, ViewChild} from '@angular/core';
import {RouterLink} from '@angular/router';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {
  faAddressBook,
  faArrowRightLong,
  faBox,
  faBoxesStacked,
  faCheckDouble,
  faDashboard,
  faGear,
  faHome,
  faKey,
  faLightbulb,
  faPalette,
  faServer,
  faUsers,
} from '@fortawesome/free-solid-svg-icons';
import {RequireRoleDirective} from '../../directives/required-role.directive';
import {SidebarService} from '../../services/sidebar.service';
import {buildConfig} from '../../../buildconfig';
import {FeatureFlagService} from '../../services/feature-flag.service';
import {AsyncPipe} from '@angular/common';
import {CdkConnectedOverlay, CdkOverlayOrigin} from '@angular/cdk/overlay';

@Component({
  selector: 'app-side-bar',
  standalone: true,
  templateUrl: './side-bar.component.html',
  imports: [RouterLink, FaIconComponent, RequireRoleDirective, AsyncPipe, CdkConnectedOverlay, CdkOverlayOrigin],
})
export class SideBarComponent {
  public readonly sidebar = inject(SidebarService);
  public readonly featureFlags = inject(FeatureFlagService);
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
  protected readonly buildConfig = buildConfig;

  @ViewChild('asideElement') private asideElement?: ElementRef<HTMLElement>;

  constructor() {
    effect(() => {
      const show = this.sidebar.showSidebar();
      this.asideElement?.nativeElement.classList.toggle('translate-x-0', show);
      this.asideElement?.nativeElement.classList.toggle('-translate-x-full', !show);
    });
  }

  protected readonly faArrowRightLong = faArrowRightLong;
  protected readonly faHome = faHome;
  protected showRequestAccessTooltip = false;
}
