import {Component, inject, ViewChild} from '@angular/core';
import {ApexOptions, ChartComponent, NgApexchartsModule} from 'ng-apexcharts';
import {DeploymentTargetsService} from '../../../services/deployment-targets.service';
import {map} from 'rxjs';
import {UserAccountWithRole} from '../../../types/user-account';

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

  private readonly deploymentTargets = inject(DeploymentTargetsService);
  private readonly deploymentTargets$ = this.deploymentTargets.list();

  constructor() {
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
      title: {
        text: 'Deployment Managers',
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
    this.deploymentTargets$.subscribe((dts) => {
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
      this.chartOptions.labels = Object.values(managers).map((v) => v.name || v.email);
      this.chartOptions.series = Object.values(counts);
    });
  }
}
