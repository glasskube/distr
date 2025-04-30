import {AsyncPipe} from '@angular/common';
import {Component, inject} from '@angular/core';
import {DeploymentTargetsComponent} from '../../deployments/deployment-targets.component';
import {ArtifactsByCustomerCardComponent} from '../../artifacts/artifacts-by-customer-card/artifacts-by-customer-card.component';
import {DashboardService} from '../../services/dashboard.service';

@Component({
  selector: 'app-dashboard',
  imports: [DeploymentTargetsComponent, AsyncPipe, ArtifactsByCustomerCardComponent],
  templateUrl: './dashboard.component.html',
})
export class DashboardComponent {
  private readonly dashboardService = inject(DashboardService);
  protected readonly artifactsByCustomer$ = this.dashboardService.getArtifactsByCustomer();
}
