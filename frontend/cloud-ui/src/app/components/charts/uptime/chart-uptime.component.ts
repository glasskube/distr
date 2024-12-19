import {Component, inject, OnInit, ViewChild} from '@angular/core';
import {ApexOptions, ChartComponent, NgApexchartsModule} from 'ng-apexcharts';
import {MetricsService} from '../../../services/metrics.service';
import {DeploymentTargetsService} from '../../../services/deployment-targets.service';
import {EMPTY, filter, first, firstValueFrom, lastValueFrom, map, switchMap} from 'rxjs';

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
export class ChartUptimeComponent implements OnInit {
  @ViewChild('chart') chart!: ChartComponent;
  public chartOptions?: ApexOptions;

  private readonly deploymentTargets = inject(DeploymentTargetsService);
  private readonly deploymentTargets$ = this.deploymentTargets.list();

  private readonly metrics = inject(MetricsService);

  constructor() {}

  async ngOnInit() {
    const dts = await firstValueFrom(this.deploymentTargets$);
    for (const dt of dts) {
      if (dt.currentStatus) {
        let deployment;
        try {
          // temporarily: simply show uptime of the first deployment target with a status and a deployment
          deployment = await lastValueFrom(this.deploymentTargets.latestDeploymentFor(dt.id!));
        } catch (e) {
          continue;
        }
        this.metrics.getUptimeForDeployment(deployment.id!).subscribe((uptimes) => {
          this.chartOptions = {
            series: [
              {
                name: 'available',
                data: uptimes.map((ut) => ut.total - ut.unknown),
                color: '#00bfa5',
              },
              {
                name: 'unknown',
                data: uptimes.map((ut) => ut.unknown),
                color: '#feb019',
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
              categories: uptimes.map((ut) => ut.hour),
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
              text: `${dt.name}`,
              align: 'center',
              style: {
                color: 'rgb(156, 163, 175)',
                fontFamily: 'Poppins',
              },
            },
          };
        });
        return;
      }
    }
  }
}
