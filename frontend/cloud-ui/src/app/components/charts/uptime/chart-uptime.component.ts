import {Component, inject, OnInit} from '@angular/core';
import {ApexOptions, NgApexchartsModule} from 'ng-apexcharts';
import {catchError, EMPTY, filter, firstValueFrom, lastValueFrom, of, switchMap} from 'rxjs';
import {DeploymentTargetsService} from '../../../services/deployment-targets.service';
import {MetricsService} from '../../../services/metrics.service';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faEllipsis, faEllipsisVertical} from '@fortawesome/free-solid-svg-icons';
import {CdkConnectedOverlay, CdkOverlayOrigin} from '@angular/cdk/overlay';
import {dropdownAnimation} from '../../../animations/dropdown';
import {AsyncPipe} from '@angular/common';
import {DeploymentTarget} from '../../../types/deployment-target';
import {HttpErrorResponse} from '@angular/common/http';
import {Deployment} from '../../../types/deployment';

@Component({
  selector: 'app-chart-uptime',
  templateUrl: './chart-uptime.component.html',
  imports: [NgApexchartsModule, FaIconComponent, CdkOverlayOrigin, CdkConnectedOverlay, AsyncPipe],
  animations: [dropdownAnimation],
})
export class ChartUptimeComponent implements OnInit {
  private readonly LOCAL_STORAGE_KEY = 'dashboard_uptime_deployment_target_id';
  public chartOptions?: ApexOptions;

  private readonly deploymentTargets = inject(DeploymentTargetsService);
  readonly deploymentTargets$ = this.deploymentTargets.list();

  private readonly metrics = inject(MetricsService);

  showDropdown = false;
  storedDeploymentTargetId?: string;
  selectedDeploymentTarget?: DeploymentTarget;

  protected readonly faEllipsis = faEllipsis;

  async ngOnInit() {
    this.storedDeploymentTargetId = window.localStorage[this.LOCAL_STORAGE_KEY];
    const dts = await firstValueFrom(this.deploymentTargets$);
    const stored = dts.find((dt) => dt.id === this.storedDeploymentTargetId);
    if (stored) {
      this.selectDeploymentTarget(stored);
    }
  }

  selectDeploymentTarget(dt: DeploymentTarget) {
    this.showDropdown = false;
    this.selectedDeploymentTarget = dt;
    window.localStorage[this.LOCAL_STORAGE_KEY] = dt.id;
    this.chartOptions = undefined;
    this.deploymentTargets
      .latestDeploymentFor(dt.id!)
      .pipe(
        catchError(() => EMPTY),
        switchMap((deployment) => (deployment?.id ? this.metrics.getUptimeForDeployment(deployment.id) : of([])))
      )
      .subscribe((uptimes) => {
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
  }
}
