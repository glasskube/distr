import {Component, inject} from '@angular/core';
import {GlobeComponent} from '../globe/globe.component';
import {DeploymentTargetsService} from '../../services/deployment-targets.service';
import {ApplicationsComponent} from '../../applications/applications.component';
import {DeploymentTargetsComponent} from '../../deployment-targets/deployment-targets.component';

@Component({
  selector: 'app-dashboard-placeholder',
  standalone: true,
  imports: [ApplicationsComponent, DeploymentTargetsComponent, GlobeComponent, AsyncPipe],
  templateUrl: './dashboard-placeholder.component.html',
})
export class DashboardPlaceholderComponent {
  private readonly deploymentTargets = inject(DeploymentTargetsService);
  public readonly deploymentTargets$ = this.deploymentTargets.list();
}
