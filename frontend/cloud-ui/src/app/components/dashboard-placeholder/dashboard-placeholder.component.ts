import {Component} from '@angular/core';
import {ApplicationsComponent} from '../../applications/applications.component';
import {DeploymentTargetsComponent} from '../../deployment-targets/deployment-targets.component';

@Component({
  selector: 'app-dashboard-placeholder',
  standalone: true,
  templateUrl: './dashboard-placeholder.component.html',
  imports: [ApplicationsComponent, DeploymentTargetsComponent],
})
export class DashboardPlaceholderComponent {}
