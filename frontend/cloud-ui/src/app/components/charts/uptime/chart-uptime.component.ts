import {Component, ViewChild} from '@angular/core';
import {ApexOptions, ChartComponent, NgApexchartsModule} from 'ng-apexcharts';

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

  constructor() {
    this.chartOptions = {
      series: [
        {
          name: 'available',
          data: [10, 10, 9, 10, 9, 10, 10],
          color: '#00bfa5',
        },
        {
          name: 'unavailable',
          data: [0, 0, 1, 0, 1, 0, 0],
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
          '2018-09-19T00:00:00.000Z',
          '2018-09-19T01:30:00.000Z',
          '2018-09-19T02:30:00.000Z',
          '2018-09-19T03:30:00.000Z',
          '2018-09-19T04:30:00.000Z',
          '2018-09-19T05:30:00.000Z',
          '2018-09-19T06:30:00.000Z',
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
