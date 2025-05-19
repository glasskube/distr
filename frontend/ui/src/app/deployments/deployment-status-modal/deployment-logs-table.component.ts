import {Component, effect, inject, input} from '@angular/core';
import {DeploymentLogsService} from '../../services/deployment-logs.service';
import {delay, map, Observable, of, scan, startWith, Subject, switchMap, tap} from 'rxjs';
import {toObservable} from '@angular/core/rxjs-interop';
import {DeploymentLogRecord} from '../../types/deployment-log-record';
import {DeploymentStatusTableComponent, DeploymentStatusTableEntry} from './deployment-status-table.component';
import {AsyncPipe} from '@angular/common';

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
            (click)="showMore()">
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

  private currentOldest?: Date;
  protected hasMore = true;
  private readonly showMore$ = new Subject<void>();

  public readonly deploymentId = input.required<string>();
  public readonly resource = input.required<string>();
  protected readonly logs$: Observable<DeploymentStatusTableEntry[]> = toObservable(this.resource).pipe(
    delay(0), // needed because this has to run after the effect() below
    switchMap((resource) =>
      resource
        ? this.showMore$.pipe(
            startWith(undefined),
            switchMap(() => this.svc.get(this.deploymentId(), resource, this.currentOldest)),
            tap((logs) => (this.hasMore = logs.length > 0)),
            scan((acc: DeploymentLogRecord[], logs) => (acc ?? []).concat(logs)),
            tap((logs) => {
              logs.forEach((rec) => {
                const ts = new Date(rec.timestamp);
                if (!this.currentOldest || ts < this.currentOldest) {
                  this.currentOldest = ts;
                }
              });
            })
          )
        : of([])
    ),
    map((logs) => logs.map((it) => ({date: it.timestamp, status: it.severity, detail: it.body})))
  );

  constructor() {
    effect(() => {
      this.resource();
      this.currentOldest = undefined;
      this.hasMore = true;
    });
  }

  protected showMore() {
    this.showMore$.next();
  }
}
