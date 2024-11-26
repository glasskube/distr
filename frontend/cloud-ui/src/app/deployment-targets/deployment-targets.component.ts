import {Component, inject, signal} from '@angular/core';
import {DeploymentTargetsService} from '../services/deployment-targets.service';
import {AsyncPipe, JsonPipe} from '@angular/common';
import {JsonpClientBackend} from '@angular/common/http';

@Component({
  selector: 'app-deployment-targets',
  imports: [AsyncPipe, JsonPipe],
  templateUrl: './deployment-targets.component.html',
})
export class DeploymentTargetsComponent {
  private readonly deploymentTargets = inject(DeploymentTargetsService);
  public deploymentTargets$ = this.deploymentTargets.list();
}
