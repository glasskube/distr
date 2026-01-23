import {AsyncPipe} from '@angular/common';
import {Component, inject, input, output, signal} from '@angular/core';
import {toObservable} from '@angular/core/rxjs-interop';
import {DeploymentTarget, DeploymentWithLatestRevision} from '@distr-sh/distr-sdk';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faXmark} from '@fortawesome/free-solid-svg-icons';
import {catchError, distinctUntilChanged, EMPTY, filter, map, switchMap, timer} from 'rxjs';
import {DeploymentLogsService} from '../../services/deployment-logs.service';
import {DeploymentLogsTableComponent} from './deployment-logs-table.component';
import {DeploymentStatusTableComponent} from './deployment-status-table.component';

const resourceRefreshInterval = 15_000;

@Component({
  selector: 'app-deployment-status-modal',
  templateUrl: './deployment-status-modal.component.html',
  imports: [AsyncPipe, DeploymentLogsTableComponent, DeploymentStatusTableComponent, FaIconComponent],
})
export class DeploymentStatusModalComponent {
  public readonly deploymentTarget = input.required<DeploymentTarget>();
  public readonly deployment = input.required<DeploymentWithLatestRevision>();
  public readonly closed = output<void>();

  protected readonly faXmark = faXmark;

  private readonly deploymentLogs = inject(DeploymentLogsService);

  private readonly deploymentId$ = toObservable(this.deployment).pipe(
    map((d) => d.id),
    filter((id) => id !== undefined),
    distinctUntilChanged()
  );

  protected readonly resources = this.deploymentId$.pipe(
    switchMap((id) =>
      timer(0, resourceRefreshInterval).pipe(
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
