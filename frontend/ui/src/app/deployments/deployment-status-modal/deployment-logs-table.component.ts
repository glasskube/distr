import {Component, computed, inject, input} from '@angular/core';
import {map, Observable} from 'rxjs';
import {DeploymentLogsService} from '../../services/deployment-logs.service';
import {DeploymentLogRecord} from '../../types/deployment-log-record';
import {TimeseriesTableComponent, TimeseriesEntry, TimeseriesSource} from './timeseries-table.component';

function logRecordToTimeseriesEntry(record: DeploymentLogRecord): TimeseriesEntry {
  return {date: record.timestamp, status: record.severity, detail: record.body};
}

class LogsTimeseriesSource implements TimeseriesSource {
  public readonly batchSize = 25;

  constructor(
    private readonly svc: DeploymentLogsService,
    private readonly deploymentId: string,
    private readonly resource: string
  ) {}

  load(): Observable<TimeseriesEntry[]> {
    return this.svc
      .get(this.deploymentId, this.resource, {limit: this.batchSize})
      .pipe(map((logs) => logs.map(logRecordToTimeseriesEntry)));
  }

  loadAfter(after: Date): Observable<TimeseriesEntry[]> {
    // TODO: Handle after
    return this.svc
      .get(this.deploymentId, this.resource, {limit: this.batchSize, after})
      .pipe(map((logs) => logs.map(logRecordToTimeseriesEntry)));
  }

  loadBefore(before: Date): Observable<TimeseriesEntry[]> {
    return this.svc
      .get(this.deploymentId, this.resource, {limit: this.batchSize, before})
      .pipe(map((logs) => logs.map(logRecordToTimeseriesEntry)));
  }
}

@Component({
  selector: 'app-deployment-logs-table',
  template: `<app-timeseries-table [source]="this.source()" />`,
  imports: [TimeseriesTableComponent],
})
export class DeploymentLogsTableComponent {
  private readonly svc = inject(DeploymentLogsService);
  public readonly deploymentId = input.required<string>();
  public readonly resource = input.required<string>();
  protected readonly source = computed<TimeseriesSource>(
    () => new LogsTimeseriesSource(this.svc, this.deploymentId(), this.resource())
  );
}
