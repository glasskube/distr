import {Component, computed, inject, input} from '@angular/core';
import {DeploymentRevisionStatus} from '@glasskube/distr-sdk';
import {map, Observable} from 'rxjs';
import {DeploymentStatusService} from '../../services/deployment-status.service';
import {TimeseriesEntry, TimeseriesSource, TimeseriesTableComponent} from './timeseries-table.component';

function statusToTimeseriesEntry(record: DeploymentRevisionStatus): TimeseriesEntry {
  return {id: record.id, date: record.createdAt!, status: record.type, detail: record.message};
}

class LogsTimeseriesSource implements TimeseriesSource {
  public readonly batchSize = 25;

  constructor(
    private readonly svc: DeploymentStatusService,
    private readonly deploymentId: string
  ) {}

  load(): Observable<TimeseriesEntry[]> {
    return this.svc
      .getStatuses(this.deploymentId, {limit: this.batchSize})
      .pipe(map((logs) => logs.map(statusToTimeseriesEntry)));
  }

  loadAfter(after: Date): Observable<TimeseriesEntry[]> {
    return this.svc
      .getStatuses(this.deploymentId, {limit: this.batchSize, after})
      .pipe(map((logs) => logs.map(statusToTimeseriesEntry)));
  }

  loadBefore(before: Date): Observable<TimeseriesEntry[]> {
    return this.svc
      .getStatuses(this.deploymentId, {limit: this.batchSize, before})
      .pipe(map((logs) => logs.map(statusToTimeseriesEntry)));
  }
}

@Component({
  selector: 'app-deployment-status-table',
  template: `<app-timeseries-table [source]="this.source()" />`,
  imports: [TimeseriesTableComponent],
})
export class DeploymentStatusTableComponent {
  private readonly svc = inject(DeploymentStatusService);
  public readonly deploymentId = input.required<string>();
  protected readonly source = computed<TimeseriesSource>(() => new LogsTimeseriesSource(this.svc, this.deploymentId()));
}
