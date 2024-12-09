import {Component} from '@angular/core';
import {NgIf} from '@angular/common';
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

@Component({
  selector: 'app-side-bar',
  standalone: true,
  templateUrl: './side-bar.component.html',
  imports: [NgIf, RouterLink, FaIconComponent],
})
export class SideBarComponent {
  public feedbackAlert = true;
  protected readonly faDashboard = faDashboard;
  protected readonly faBoxesStacked = faBoxesStacked;
  protected readonly faServer = faServer;
  protected readonly faLightbulb = faLightbulb;
  protected readonly faKey = faKey;
  protected readonly faGear = faGear;
  protected readonly faCheckDouble = faCheckDouble;
  protected readonly faPalette = faPalette;

  protected readonly userRole: 'distributor' | 'customer' = 'customer'; // TODO: load from auth service
}
