import {Component, Input} from '@angular/core';
import {DeploymentTarget} from '../types/deployment-target';
import {IsStalePipe} from '../../util/model';

@Component({
  selector: 'app-status-dot',
  template: `
    <div
      class="rounded-full w-full h-full"
      [class.bg-lime-600]="deploymentTarget.currentStatus && !(deploymentTarget.currentStatus | isStale)"
      [class.bg-yellow-300]="deploymentTarget.currentStatus && (deploymentTarget.currentStatus | isStale)"
      [class.bg-gray-500]="!deploymentTarget.currentStatus"></div>
  `,
  imports: [IsStalePipe],
})
export class StatusDotComponent {
  @Input({required: true}) deploymentTarget!: DeploymentTarget;
}
