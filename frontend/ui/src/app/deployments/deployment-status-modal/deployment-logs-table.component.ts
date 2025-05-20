import {AsyncPipe} from '@angular/common';
import {Component, inject, input} from '@angular/core';
import {toObservable} from '@angular/core/rxjs-interop';
import {map, Observable, of, scan, startWith, Subject, switchMap, tap} from 'rxjs';
import {DeploymentLogsService} from '../../services/deployment-logs.service';
import {DeploymentLogRecord} from '../../types/deployment-log-record';
import {DeploymentStatusTableComponent, DeploymentStatusTableEntry} from './deployment-status-table.component';

@Component({
  selector: 'app-deployment-logs-table',
  template: `
    @if (logs$ | async; as logs) {
      <app-deployment-status-table [entries]="logs"></app-deployment-status-table>
      @if (hasMore) {
        <div class="flex items-center justify-center mt-2">
          <button
            type="button"
            class="py-2 px-3 flex items-center text-sm font-medium text-center text-gray-900 focus:outline-none bg-white rounded-lg border border-gray-200 hover:bg-gray-100 hover:text-primary-700 focus:z-10 focus:ring-4 focus:ring-gray-200 dark:focus:ring-gray-700 dark:bg-gray-800 dark:text-gray-400 dark:border-gray-600 dark:hover:text-white dark:hover:bg-gray-700"
            (click)="showMore(logs)">
            Load more
          </button>
        </div>
      }
    }
  `,
  imports: [DeploymentStatusTableComponent, AsyncPipe],
})
export class DeploymentLogsTableComponent {
  private readonly svc = inject(DeploymentLogsService);

  protected hasMore = true;
  private readonly showMoreBefore$ = new Subject<Date>();

  public readonly deploymentId = input.required<string>();
  public readonly resource = input.required<string>();
  protected readonly logs$: Observable<DeploymentStatusTableEntry[]> = toObservable(this.resource).pipe(
    switchMap((resource) =>
      resource
        ? this.showMoreBefore$.pipe(
            startWith(undefined),
            switchMap((before) => this.svc.get(this.deploymentId(), resource, before)),
            tap((logs) => (this.hasMore = logs.length > 0)),
            scan((acc: DeploymentLogRecord[], logs) => (acc ?? []).concat(logs))
          )
        : of([])
    ),
    map((logs) => logs.map((it) => ({date: it.timestamp, status: it.severity, detail: it.body})))
  );

  protected showMore(logs: DeploymentStatusTableEntry[]) {
    let before: Date | null = null;
    logs.forEach((rec) => {
      const d = new Date(rec.date);
      if (before === null || d < before) {
        before = d;
      }
    });
    if (before !== null) {
      this.showMoreBefore$.next(before);
    }
  }
}
