import {AsyncPipe} from '@angular/common';
import {Component, inject, OnDestroy, OnInit} from '@angular/core';
import {DeploymentTargetsComponent} from '../../deployments/deployment-targets.component';
import {ArtifactsByCustomerCardComponent} from '../../artifacts/artifacts-by-customer-card/artifacts-by-customer-card.component';
import {DashboardService} from '../../services/dashboard.service';
import {ActivatedRoute, Router} from '@angular/router';
import {combineLatestWith, first, Subject, switchMap, takeUntil} from 'rxjs';
import {DeploymentTargetsService} from '../../services/deployment-targets.service';

@Component({
  selector: 'app-dashboard',
  imports: [DeploymentTargetsComponent, AsyncPipe, ArtifactsByCustomerCardComponent],
  templateUrl: './dashboard.component.html',
})
export class DashboardComponent implements OnInit, OnDestroy {
  private readonly route = inject(ActivatedRoute);
  private readonly router = inject(Router);
  private readonly dashboardService = inject(DashboardService);
  protected readonly artifactsByCustomer$ = this.dashboardService.getArtifactsByCustomer();
  private readonly deploymentTargetsService = inject(DeploymentTargetsService);
  private readonly destroyed$ = new Subject<void>();

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
              return this.router.navigate([this.router.url]);
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
