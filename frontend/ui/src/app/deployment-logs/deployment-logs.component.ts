import {Component, effect, forwardRef, inject, input, output, resource, signal} from '@angular/core';
import {firstValueFrom} from 'rxjs';
import {DeploymentLogsService} from '../services/deployment-logs.service';
import {DeploymentResourceLogsComponent} from './deployment-resource-logs.component';
import {Deployment, DeploymentWithLatestRevision} from '@glasskube/distr-sdk';

@Component({
  selector: 'app-deployment-logs',
  templateUrl: './deployment-logs.component.html',
  imports: [DeploymentResourceLogsComponent],
})
export class DeploymentLogsComponent {
  private readonly svc = inject(DeploymentLogsService);

  public readonly deployment = input.required<DeploymentWithLatestRevision>();
  public readonly closed = output<void>();
  protected readonly resources = resource({
    request: () => ({deploymentId: this.deployment().id}),
    loader: (param) => firstValueFrom(this.svc.getResources(param.request.deploymentId!)),
  });
  protected readonly selectedResource = signal<string | null>(null);

  constructor() {
    effect(() => {
      const resources = this.resources.value();
      this.selectedResource.set(resources?.length ? resources[0] : null);
    });
  }
}
