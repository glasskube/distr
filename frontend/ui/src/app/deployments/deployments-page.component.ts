import {Component} from '@angular/core';
import {DeploymentTargetsComponent} from './deployment-targets.component';

@Component({
  selector: 'app-deployments-page',
  imports: [DeploymentTargetsComponent],
  templateUrl: './deployments-page.component.html',
})
export class DeploymentsPageComponent {}
