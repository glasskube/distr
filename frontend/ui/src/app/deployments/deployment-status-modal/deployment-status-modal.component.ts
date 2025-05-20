import {Component, inject, input, output, signal} from '@angular/core';
import {toObservable} from '@angular/core/rxjs-interop';
import {DeploymentTarget, DeploymentWithLatestRevision} from '@glasskube/distr-sdk';
import {catchError, distinctUntilChanged, EMPTY, filter, interval, map, Observable, switchMap, timer} from 'rxjs';
import {DeploymentStatusService} from '../../services/deployment-status.service';
import {AsyncPipe} from '@angular/common';
import {DeploymentLogsService} from '../../services/deployment-logs.service';
import {DeploymentStatusTableComponent, DeploymentStatusTableEntry} from './deployment-status-table.component';
import {DeploymentLogsTableComponent} from './deployment-logs-table.component';

const resourceRefreshInterval = 15_000;

@Component({
  selector: 'app-deployment-status-modal',
  templateUrl: './deployment-status-modal.component.html',
  imports: [AsyncPipe, DeploymentStatusTableComponent, DeploymentLogsTableComponent],
})
export class DeploymentStatusModalComponent {
  private readonly deploymentStatuses = inject(DeploymentStatusService);
  private readonly deploymentLogs = inject(DeploymentLogsService);

  public readonly deploymentTarget = input.required<DeploymentTarget>();
  public readonly selectedDeployment = input.required<DeploymentWithLatestRevision>();
  public readonly closed = output<void>();

  private readonly deploymentID$ = toObservable(this.selectedDeployment).pipe(
    map((d) => d.id),
    filter((id) => id !== undefined),
    distinctUntilChanged()
  );

  protected readonly statuses: Observable<DeploymentStatusTableEntry[]> = this.deploymentID$.pipe(
    switchMap((id) => this.deploymentStatuses.pollStatuses(id)),
    map((statuses) => statuses.map((it) => ({id: it.id, date: it.createdAt!, status: it.type, detail: it.message})))
  );
  protected readonly resources = this.deploymentID$.pipe(
    switchMap((id) =>
      interval(resourceRefreshInterval).pipe(
        switchMap(() => this.deploymentLogs.getResources(id).pipe(catchError(() => EMPTY)))
      )
    )
  );

  /**
   * `null` means agent status
   */
  protected readonly selectedResource = signal<string | null>(null);

  protected hideModal() {
    this.closed.emit();
  }
}
