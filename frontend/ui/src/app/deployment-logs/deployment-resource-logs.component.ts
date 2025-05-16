import {AsyncPipe, DatePipe} from '@angular/common';
import {Component, inject, input, effect} from '@angular/core';
import {toObservable} from '@angular/core/rxjs-interop';
import {Subject, delay, switchMap, startWith, tap, scan, of} from 'rxjs';
import {DeploymentLogsService} from '../services/deployment-logs.service';
import {DeploymentLogRecord} from '../types/deployment-log-record';

@Component({
  selector: 'app-deployment-resource-logs',
  templateUrl: './deployment-resource-logs.component.html',
  imports: [AsyncPipe, DatePipe],
})
export class DeploymentResourceLogsComponent {
  private readonly svc = inject(DeploymentLogsService);

  private currentOldest?: Date;
  protected hasMore = true;
  private readonly showMore$ = new Subject<void>();

  public readonly deploymentId = input.required<string>();
  public readonly resource = input.required<string>();
  protected readonly logs = toObservable(this.resource).pipe(
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
                if (!this.currentOldest || new Date(rec.timestamp) < this.currentOldest) {
                  this.currentOldest = ts;
                }
              });
            })
          )
        : of([])
    )
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
