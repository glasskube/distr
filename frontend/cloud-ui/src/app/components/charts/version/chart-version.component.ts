import {Component, ViewChild} from '@angular/core';
import {ApexOptions, ChartComponent, NgApexchartsModule} from 'ng-apexcharts';

@Component({
  selector: 'app-chart-version',
  templateUrl: './chart-version.component.html',
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
export class ChartVersionComponent {
  @ViewChild('chart') chart!: ChartComponent;
  public chartOptions: ApexOptions;

  constructor() {
    this.chartOptions = {
      series: [
        {
          name: 'v1.0.0',
          data: [2, 2, 2, 2, 1, 1, 1],
        },
        {
          name: 'v4.2.0',
          data: [0, 1, 2, 3, 4, 3, 1],
        },
        {
          name: 'v4.2.1',
          data: [0, 0, 0, 0, 1, 2, 7],
        },
      ],
      chart: {
        height: 192,

        type: 'area',
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
      legend: {
        show: true,
        position: 'top',
        fontFamily: 'Inter',
        labels: {
          colors: 'rgb(156, 163, 175)',
          useSeriesColors: false,
        },
      },
      title: {
        text: 'Version Distribution',
        align: 'center',
        style: {
          color: 'rgb(156, 163, 175)',
          fontFamily: 'Poppins',
        },
      },
    };
  }
}
