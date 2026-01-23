import {Component, computed, inject, input} from '@angular/core';
import {map, Observable} from 'rxjs';
import {
  TimeseriesEntry,
  TimeseriesExporter,
  TimeseriesSource,
  TimeseriesTableComponent,
} from '../../components/timeseries-table.component';
import {DeploymentTargetLogsService} from '../../services/deployment-target-logs.service';
import {DeploymentTargetLogRecord} from '../../types/deployment-target-log-record';

function logRecordToTimeseriesEntry(record: DeploymentTargetLogRecord): TimeseriesEntry {
  return {id: record.id, date: record.timestamp, status: record.severity, detail: record.body.trim()};
}

class LogsTimeseriesSource implements TimeseriesSource {
  public readonly batchSize = 25;

  constructor(
    private readonly svc: DeploymentTargetLogsService,
    private readonly deploymentTargetId: string
  ) {}

  load(): Observable<TimeseriesEntry[]> {
    return this.svc
      .get(this.deploymentTargetId, {limit: this.batchSize})
      .pipe(map((logs) => logs.map(logRecordToTimeseriesEntry)));
  }

  loadAfter(after: Date): Observable<TimeseriesEntry[]> {
    return this.svc
      .get(this.deploymentTargetId, {limit: this.batchSize, after})
      .pipe(map((logs) => logs.map(logRecordToTimeseriesEntry)));
  }

  loadBefore(before: Date): Observable<TimeseriesEntry[]> {
    return this.svc
      .get(this.deploymentTargetId, {limit: this.batchSize, before})
      .pipe(map((logs) => logs.map(logRecordToTimeseriesEntry)));
  }
}

@Component({
  selector: 'app-deployment-target-logs-table',
  template: `<app-timeseries-table [source]="source()" [exporter]="exporter" />`,
  imports: [TimeseriesTableComponent],
})
export class DeploymentTargetLogsTableComponent {
  private readonly svc = inject(DeploymentTargetLogsService);
  public readonly deploymentTargetId = input.required<string>();
  protected readonly source = computed<TimeseriesSource>(
    () => new LogsTimeseriesSource(this.svc, this.deploymentTargetId())
  );
  protected readonly exporter: TimeseriesExporter = {
    export: () => this.svc.export(this.deploymentTargetId()),
    getFileName: () => 'agent.log',
  };
}
