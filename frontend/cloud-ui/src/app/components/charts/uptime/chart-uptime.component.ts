import {Component, inject, ViewChild} from '@angular/core';
import {ApexOptions, ChartComponent, NgApexchartsModule} from 'ng-apexcharts';
import {MetricsService} from '../../../services/metrics.service';
import {DeploymentTargetsService} from '../../../services/deployment-targets.service';
import {switchMap} from 'rxjs';

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
    this.metrics.getUptimeForDeployment('11a676ea-714a-4e5e-9a67-54c9d2f7b64c').subscribe(console.log);
    this.chartOptions = {
      series: [
        {
          name: 'available',
          data: [10, 10, 9, 10, 9, 10, 10, 10, 10, 9, 10, 9, 10, 10, 10, 10, 9, 10, 9, 10, 10, 10, 10, 10],
          color: '#00bfa5',
        },
        {
          name: 'unavailable',
          data: [0, 0, 1, 0, 1, 0, 0, 0, 0, 1, 0, 1, 0, 0, 0, 0, 1, 0, 1, 0, 0, 0, 0, 0],
          color: '#f44336',
        },
      ],
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
          '2018-09-19T00:30:00.000Z',
          '2018-09-19T01:30:00.000Z',
          '2018-09-19T02:30:00.000Z',
          '2018-09-19T03:30:00.000Z',
          '2018-09-19T04:30:00.000Z',
          '2018-09-19T05:30:00.000Z',
          '2018-09-19T06:30:00.000Z',
          '2018-09-19T07:30:00.000Z',
          '2018-09-19T08:30:00.000Z',
          '2018-09-19T09:30:00.000Z',
          '2018-09-19T10:30:00.000Z',
          '2018-09-19T11:30:00.000Z',
          '2018-09-19T12:30:00.000Z',
          '2018-09-19T13:30:00.000Z',
          '2018-09-19T14:30:00.000Z',
          '2018-09-19T15:30:00.000Z',
          '2018-09-19T16:30:00.000Z',
          '2018-09-19T17:30:00.000Z',
          '2018-09-19T18:30:00.000Z',
          '2018-09-19T19:30:00.000Z',
          '2018-09-19T20:30:00.000Z',
          '2018-09-19T21:30:00.000Z',
          '2018-09-19T22:30:00.000Z',
          '2018-09-19T23:30:00.000Z',
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
  }
}
