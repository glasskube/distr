import {AsyncPipe} from '@angular/common';
import {Component, inject, OnDestroy, OnInit} from '@angular/core';
import {ArtifactsByCustomerCardComponent} from '../../artifacts/artifacts-by-customer-card/artifacts-by-customer-card.component';
import {DashboardService} from '../../services/dashboard.service';
import {ActivatedRoute, Router} from '@angular/router';
import {combineLatestWith, first, shareReplay, Subject, switchMap, takeUntil} from 'rxjs';
import {DeploymentTargetsService} from '../../services/deployment-targets.service';
import {DeploymentTargetCardComponent} from '../../deployments/deployment-target-card/deployment-target-card.component';

@Component({
  selector: 'app-dashboard',
  imports: [AsyncPipe, ArtifactsByCustomerCardComponent, DeploymentTargetCardComponent],
  templateUrl: './dashboard.component.html',
})
export class DashboardComponent implements OnInit, OnDestroy {
  private readonly destroyed$ = new Subject<void>();
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly dashboardService = inject(DashboardService);
  protected readonly artifactsByCustomer$ = this.dashboardService.getArtifactsByCustomer().pipe(shareReplay(1));
  private readonly deploymentTargetsService = inject(DeploymentTargetsService);
  protected readonly deploymentTargets$ = this.deploymentTargetsService
    .poll()
    .pipe(takeUntil(this.destroyed$), shareReplay(1));

  ngOnInit() {
    if (this.route.snapshot.queryParams?.['from'] === 'login') {
      this.artifactsByCustomer$
        .pipe(
          takeUntil(this.destroyed$),
          combineLatestWith(this.deploymentTargetsService.list()),
          first(),
          switchMap(([artifacts, dts]) => {
            if (artifacts.length === 0 && dts.length === 0) {
              return this.router.navigate(['tutorials']);
            } else {
              return this.router.navigate([this.router.url]); // remove query param
            }
          })
        )
        .subscribe();
    }
  }

  ngOnDestroy() {
    this.destroyed$.next();
    this.destroyed$.complete();
  }
}
