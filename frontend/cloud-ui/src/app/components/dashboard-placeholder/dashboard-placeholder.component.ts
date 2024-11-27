import {Component} from '@angular/core';
import {ApplicationsComponent} from '../../applications/applications.component';

@Component({
  selector: 'app-dashboard-placeholder',
  standalone: true,
  templateUrl: './dashboard-placeholder.component.html',
  imports: [ApplicationsComponent],
})
export class DashboardPlaceholderComponent {}
