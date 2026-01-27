import {NgClass} from '@angular/common';
import {Component, computed, Directive, input, Signal} from '@angular/core';
import {DeploymentStatusType, DeploymentTarget, DeploymentWithLatestRevision} from '@distr-sh/distr-sdk';
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

@Directive({
  selector: '[appDeploymentStatusDot]',
  host: {
    '[class.rounded-full]': 'true',
    '[class.bg-gray-500]': 'isInitial()',
    '[class.bg-red-400]': 'isError()',
    '[class.bg-yellow-300]': 'isStale()',
    '[class.bg-blue-400]': 'isProgressing()',
    '[class.border]': 'isRunning()',
    '[class.border-3]': 'isRunning()',
    '[class.border-lime-600]': 'isRunning()',
    '[class.bg-lime-600]': 'isHealthy()',
  },
})
export class DeploymentStatusDotDirective {
  public readonly deployment = input.required<DeploymentWithLatestRevision>();

  protected readonly isInitial = computed(() => this.deployment().latestStatus === undefined);
  protected readonly isStale = computed(() => {
    const status = this.deployment().latestStatus;
    return status !== undefined && status.type !== 'error' && isStale(status);
  });
  protected readonly isError = this.statusSignal('error', true);
  protected readonly isProgressing = this.statusSignal('progressing');
  protected readonly isRunning = this.statusSignal('running');
  protected readonly isHealthy = this.statusSignal('healthy');

  private statusSignal(type: DeploymentStatusType, takePrecedenceBeforeStale?: boolean): Signal<boolean> {
    return computed(() => {
      const status = this.deployment().latestStatus;
      return status !== undefined && status.type === type && (takePrecedenceBeforeStale || !isStale(status));
    });
  }
}
