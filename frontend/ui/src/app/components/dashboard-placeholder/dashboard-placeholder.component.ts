import {AsyncPipe} from '@angular/common';
import {Component, inject} from '@angular/core';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faPlus} from '@fortawesome/free-solid-svg-icons';
import {ApplicationsComponent} from '../../applications/applications.component';
import {DeploymentTargetsComponent} from '../../deployments/deployment-targets.component';
import {ApplicationsService} from '../../services/applications.service';
import {DeploymentTargetsService} from '../../services/deployment-targets.service';
import {ChartTypeComponent} from '../charts/type/chart-type.component';
import {ChartUptimeComponent} from '../charts/uptime/chart-uptime.component';
import {GlobeComponent} from '../globe/globe.component';

@Component({
  selector: 'app-dashboard-placeholder',
  imports: [
    ApplicationsComponent,
    DeploymentTargetsComponent,
    GlobeComponent,
    AsyncPipe,
    FaIconComponent,
    ChartUptimeComponent,
    ChartTypeComponent,
  ],
  templateUrl: './dashboard-placeholder.component.html',
})
export class DashboardPlaceholderComponent {
  private readonly deploymentTargets = inject(DeploymentTargetsService);
  private readonly applications = inject(ApplicationsService);
  protected readonly deploymentTargets$ = this.deploymentTargets.list();
  protected applications$ = this.applications.list();
  protected readonly faPlus = faPlus;
}
