import {Component, inject, OnDestroy} from '@angular/core';
import {ApexOptions, NgApexchartsModule} from 'ng-apexcharts';
import {Subject, takeUntil} from 'rxjs';
import {DeploymentTargetsService} from '../../../services/deployment-targets.service';
import {UserAccountWithRole} from '@glasskube/distr-sdk';

@Component({
  selector: 'app-chart-type',
  templateUrl: './chart-type.component.html',
  imports: [NgApexchartsModule],
})
export class ChartTypeComponent implements OnDestroy {
  public chartOptions?: ApexOptions;

  private readonly deploymentTargets = inject(DeploymentTargetsService);
  private readonly deploymentTargets$ = this.deploymentTargets.list();

  private readonly destroyed$ = new Subject<void>();

  loading: boolean = true;

  constructor() {
    this.deploymentTargets$.pipe(takeUntil(this.destroyed$)).subscribe((dts) => {
      this.loading = false;

      if (dts.length > 0) {
        this.chartOptions = {
          series: [],
          labels: [],
          colors: ['#0db7ed', '#326CE5', '#174c76'],
          chart: {
            height: 192,
            type: 'donut',
            sparkline: {
              enabled: true,
            },
          },
          stroke: {
            curve: 'smooth',
          },
          tooltip: {
            enabled: false,
          },
          legend: {
            show: true,
            position: 'top',
            fontFamily: 'Inter',
            labels: {
              colors: 'rgb(156, 163, 175)',
              useSeriesColors: false,
            },
          },
          plotOptions: {
            pie: {
              customScale: 0.8,
            },
          },
        };
        const managers: {[key: string]: UserAccountWithRole} = {};
        const counts: {[key: string]: number} = {};
        for (const dt of dts) {
          if (dt.createdBy?.id && !managers[dt.createdBy.id]) {
            managers[dt.createdBy.id] = dt.createdBy;
            counts[dt.createdBy.id] = 1;
          } else if (dt.createdBy?.id && managers[dt.createdBy.id]) {
            counts[dt.createdBy.id] = counts[dt.createdBy.id] + 1;
          }
        }

        this.chartOptions.labels = Object.values(managers).map((v) => v.name ?? v.email);
        this.chartOptions.series = Object.values(counts);
      }
    });
  }

  ngOnDestroy() {
    this.destroyed$.next();
    this.destroyed$.complete();
  }
}
