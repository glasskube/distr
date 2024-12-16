import {Component, ViewChild} from '@angular/core';
import {ApexOptions, ChartComponent, NgApexchartsModule} from 'ng-apexcharts';

@Component({
  selector: 'app-chart-type',
  templateUrl: './chart-type.component.html',
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
export class ChartTypeComponent {
  @ViewChild('chart') chart!: ChartComponent;
  public chartOptions: ApexOptions;

  constructor() {
    this.chartOptions = {
      series: [4, 5, 1],
      labels: ['docker', 'kubernetes', 'glasskube'],
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
      title: {
        text: 'Deployment types (Example)',
        align: 'center',
        offsetY: 10,
        style: {
          color: 'rgb(156, 163, 175)',
          fontFamily: 'Poppins',
        },
      },
      plotOptions: {
        pie: {
          customScale: 0.8,
        },
      },
    };
  }
}
