import {Component, inject, OnInit} from '@angular/core';
import {ApexOptions, NgApexchartsModule} from 'ng-apexcharts';
import {firstValueFrom, lastValueFrom} from 'rxjs';
import {DeploymentTargetsService} from '../../../services/deployment-targets.service';
import {MetricsService} from '../../../services/metrics.service';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faEllipsis, faEllipsisVertical} from '@fortawesome/free-solid-svg-icons';
import {CdkConnectedOverlay, CdkOverlayOrigin} from '@angular/cdk/overlay';
import {dropdownAnimation} from '../../../animations/dropdown';

@Component({
  selector: 'app-chart-uptime',
  templateUrl: './chart-uptime.component.html',
  imports: [NgApexchartsModule, FaIconComponent, CdkOverlayOrigin, CdkConnectedOverlay],
  animations: [dropdownAnimation],
})
export class ChartUptimeComponent implements OnInit {
  public chartOptions?: ApexOptions;

  private readonly deploymentTargets = inject(DeploymentTargetsService);
  private readonly deploymentTargets$ = this.deploymentTargets.list();

  private readonly metrics = inject(MetricsService);

  loading = true;
  showDropdown = false;

  protected readonly faEllipsisVertical = faEllipsisVertical;
  protected readonly faEllipsis = faEllipsis;

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
          this.loading = false;
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
              offsetY: 10,
              //width: '100%',
              //height: '80%',
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
              offsetY: -5,
              floating: true,
              labels: {
                colors: 'rgb(156, 163, 175)',
                useSeriesColors: false,
              },
            },
          };
        });
        return;
      }
    }
  }

  selectDeploymentTarget() {
    // TODO
    this.showDropdown = false;
  }
}
