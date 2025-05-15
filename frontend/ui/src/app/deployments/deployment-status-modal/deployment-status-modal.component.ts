import {Component, inject, input, output} from '@angular/core';
import {toObservable} from '@angular/core/rxjs-interop';
import {DeploymentTarget, DeploymentWithLatestRevision} from '@glasskube/distr-sdk';
import {distinctUntilChanged, filter, map, switchMap} from 'rxjs';
import {DeploymentStatusService} from '../../services/deployment-status.service';
import {AsyncPipe, DatePipe} from '@angular/common';

@Component({
  selector: 'app-deployment-status-modal',
  templateUrl: './deployment-status-modal.component.html',
  imports: [AsyncPipe, DatePipe],
})
export class DeploymentStatusModalComponent {
  private readonly deploymentStatuses = inject(DeploymentStatusService);

  public readonly deploymentTarget = input.required<DeploymentTarget>();
  public readonly selectedDeployment = input.required<DeploymentWithLatestRevision>();
  public readonly closed = output<void>();

  protected readonly statuses = toObservable(this.selectedDeployment).pipe(
    map((d) => d.id),
    filter((id) => !!id),
    distinctUntilChanged(),
    switchMap((id) => this.deploymentStatuses.pollStatuses(id!))
  );

  protected hideModal() {
    this.closed.emit();
  }
}
