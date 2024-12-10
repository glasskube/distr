import {Component, effect, ElementRef, inject, OnInit, ViewChild} from '@angular/core';
import {RouterLink} from '@angular/router';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {
  faBoxesStacked,
  faCheckDouble,
  faDashboard,
  faGear,
  faKey,
  faLightbulb,
  faPalette,
  faServer,
} from '@fortawesome/free-solid-svg-icons';
import {SidebarService} from '../../services/sidebar.service';
import {AuthService} from '../../services/auth.service';

@Component({
  selector: 'app-side-bar',
  standalone: true,
  templateUrl: './side-bar.component.html',
  imports: [RouterLink, FaIconComponent],
})
export class SideBarComponent implements OnInit {
  public readonly sidebar = inject(SidebarService);
  private readonly auth = inject(AuthService);

  public feedbackAlert = true;
  protected readonly faDashboard = faDashboard;
  protected readonly faBoxesStacked = faBoxesStacked;
  protected readonly faServer = faServer;
  protected readonly faLightbulb = faLightbulb;
  protected readonly faKey = faKey;
  protected readonly faGear = faGear;
  protected readonly faCheckDouble = faCheckDouble;
  protected readonly faPalette = faPalette;

  protected userRole: 'distributor' | 'customer' = 'distributor';

  @ViewChild('asideElement') private asideElement!: ElementRef<HTMLElement>;

  constructor() {
    effect(() => {
      const show = this.sidebar.showSidebar();
      this.asideElement.nativeElement.classList.toggle('translate-x-0', show);
      this.asideElement.nativeElement.classList.toggle('-translate-x-full', !show);
    });
  }

  ngOnInit(): void {
    const {email, name} = this.auth.getClaims();
    if (email === 'pmig+customer@glasskube.com') this.userRole = 'customer';

  }
}
