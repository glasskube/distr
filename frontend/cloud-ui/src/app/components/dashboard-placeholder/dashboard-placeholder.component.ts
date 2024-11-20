import {Component, inject} from '@angular/core';
import {GlobeComponent} from '../globe/globe.component';
import {DeploymentTargetsService} from '../../services/deployment-targets.service';
import {AsyncPipe} from '@angular/common';

@Component({
  selector: 'app-dashboard-placeholder',
  standalone: true,
  imports: [GlobeComponent, AsyncPipe],
  templateUrl: './dashboard-placeholder.component.html',
})
export class DashboardPlaceholderComponent {
  private readonly deploymentTargets = inject(DeploymentTargetsService);
  public readonly deploymentTargets$ = this.deploymentTargets.list();
}
