import {Component, inject, ViewChild} from '@angular/core';
import {ApexOptions, ChartComponent, NgApexchartsModule} from 'ng-apexcharts';
import {MetricsService} from '../../../services/metrics.service';
import {DeploymentTargetsService} from '../../../services/deployment-targets.service';
import {filter, first, map, switchMap} from 'rxjs';

@Component({
  selector: 'app-chart-uptime',
  templateUrl: './chart-uptime.component.html',
  styles: [
    `
      #chart {
        max-width: 100%;
        max-height: 100%;
        margin: 0 auto;
      }
    `,
  ],
  imports: [NgApexchartsModule],
})
export class ChartUptimeComponent {
  @ViewChild('chart') chart!: ChartComponent;
  public chartOptions: ApexOptions;

  private readonly deploymentTargets = inject(DeploymentTargetsService);
  private readonly deploymentTargets$ = this.deploymentTargets.list();

  private readonly metrics = inject(MetricsService);

  constructor() {
    this.chartOptions = {
      series: [],
      chart: {
        height: 220,
        type: 'bar',
        stacked: true,
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
      xaxis: {
        type: 'datetime',
        categories: [
        ],
      },
      legend: {
        show: true,
        position: 'top',
        fontFamily: 'Inter',
        offsetY: -18,
        labels: {
          colors: 'rgb(156, 163, 175)',
          useSeriesColors: false,
        },
      },
      title: {
        text: 'Deployment Uptime (Example)',
        align: 'center',
        style: {
          color: 'rgb(156, 163, 175)',
          fontFamily: 'Poppins',
        },
      },
    };

    /*this.deploymentTargets$.pipe(
      first(),
      map(dts => dts.find(dt => dt.currentStatus)),
      filter(dt => !!dt),
      switchMap(dt => this.deploymentTargets.latestDeploymentFor(dt.id!)),
      switchMap(deployment => this.metrics.getUptimeForDeployment(deployment.id!)))*/

    this.metrics.getUptimeForDeployment('5d5e4e61-cd82-46ed-965b-6ba5d3a1d1b8').subscribe((uptimes) => {
      this.chartOptions.xaxis!.categories = uptimes.map(ut => ut.hour);
      this.chartOptions.series = [
        {
          name: 'available',
          data: uptimes.map(ut => ut.total - ut.unknown),
          color: '#00bfa5',
        },
        {
          name: 'unavailable',
          data: uptimes.map(ut => ut.unknown),
          color: '#f44336',
        },
      ]
    });
  }
}
