import {NgClass} from '@angular/common';
import {Component, computed, input} from '@angular/core';
import {DeploymentTarget, DeploymentWithLatestRevision} from '@distr-sh/distr-sdk';
import {isStale} from '../../util/model';

@Component({
  selector: 'app-status-dot',
  template: '<div [ngClass]="classList()"></div>',
  imports: [NgClass],
})
export class StatusDotComponent {
  public readonly deploymentTarget = input.required<DeploymentTarget>();

  private readonly bgClass = computed(() => {
    const dt = this.deploymentTarget();
    if (dt.currentStatus !== undefined) {
      if (isStale(dt.currentStatus)) {
        return 'bg-yellow-300';
      } else {
        return 'bg-lime-600';
      }
    }
    return 'bg-gray-500';
  });

  protected readonly classList = computed(() => ['rounded-full', 'w-full', 'h-full', this.bgClass()]);
}

@Component({
  selector: 'app-deployment-status-dot',
  template: '<div [ngClass]="classList()"></div>',
  imports: [NgClass],
})
export class DeploymentStatusDot {
  public readonly deployment = input.required<DeploymentWithLatestRevision>();

  private readonly bgClass = computed(() => {
    const d = this.deployment();
    if (d.latestStatus !== undefined) {
      if (d.latestStatus.type === 'error') {
        return 'bg-red-400';
      } else if (isStale(d.latestStatus)) {
        return 'bg-yellow-300';
      } else if (d.latestStatus.type === 'progressing') {
        return 'bg-blue-400';
      } else {
        return 'bg-lime-600';
      }
    }
    return 'bg-gray-500';
  });

  protected readonly classList = computed(() => ['rounded-full', 'w-full', 'h-full', this.bgClass()]);
}
