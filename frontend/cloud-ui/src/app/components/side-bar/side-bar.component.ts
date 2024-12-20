import {Component, effect, ElementRef, inject, ViewChild} from '@angular/core';
import {RouterLink} from '@angular/router';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {
  faAddressBook,
  faArrowRightLong,
  faBoxesStacked,
  faCheckDouble,
  faDashboard,
  faGear,
  faKey,
  faLightbulb,
  faPalette,
  faServer,
  faUsers,
} from '@fortawesome/free-solid-svg-icons';
import {RequireRoleDirective} from '../../directives/required-role.directive';
import {SidebarService} from '../../services/sidebar.service';
import {buildConfig} from '../../../buildconfig';

@Component({
  selector: 'app-side-bar',
  standalone: true,
  templateUrl: './side-bar.component.html',
  imports: [RouterLink, FaIconComponent, RequireRoleDirective],
})
export class SideBarComponent {
  public readonly sidebar = inject(SidebarService);
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
}
