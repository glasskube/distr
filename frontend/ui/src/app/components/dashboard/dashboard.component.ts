import {GlobalPositionStrategy} from '@angular/cdk/overlay';
import {AsyncPipe} from '@angular/common';
import {AfterViewInit, Component, inject, OnDestroy, TemplateRef, ViewChild} from '@angular/core';
import {faPlus} from '@fortawesome/free-solid-svg-icons';
import {combineLatest, first, Subject} from 'rxjs';
import {DeploymentTargetsComponent} from '../../deployments/deployment-targets.component';
import {ApplicationsService} from '../../services/applications.service';
import {AuthService} from '../../services/auth.service';
import {DeploymentTargetsService} from '../../services/deployment-targets.service';
import {DialogRef, OverlayService} from '../../services/overlay.service';
import {OnboardingWizardComponent} from '../onboarding-wizard/onboarding-wizard.component';
import {ArtifactsByCustomerCardComponent} from '../../artifacts/artifacts-by-customer-card/artifacts-by-customer-card.component';
import {DashboardService} from '../../services/dashboard.service';

@Component({
  selector: 'app-dashboard',
  imports: [DeploymentTargetsComponent, AsyncPipe, OnboardingWizardComponent, ArtifactsByCustomerCardComponent],
  templateUrl: './dashboard.component.html',
})
export class DashboardComponent implements AfterViewInit, OnDestroy {
  private destoryed$ = new Subject<void>();
  private overlay = inject(OverlayService);
  private readonly deploymentTargets = inject(DeploymentTargetsService);
  public readonly deploymentTargets$ = this.deploymentTargets.list();
  private readonly applications = inject(ApplicationsService);
  readonly applications$ = this.applications.list();
  private readonly auth = inject(AuthService);
  private readonly dashboardService = inject(DashboardService);

  @ViewChild('onboardingWizard') wizardRef?: TemplateRef<unknown>;

  private overlayRef?: DialogRef;

  protected readonly faPlus = faPlus;

  protected readonly artifactsByCustomer$ = this.dashboardService.getArtifactsByCustomer();

  ngAfterViewInit() {
    combineLatest([this.applications$])
      .pipe(first())
      .subscribe(([apps]) => {
        if (this.auth.hasRole('vendor') && apps.length === 0) {
          this.closeWizard();
          this.openWizard();
        }
      });
  }

  ngOnDestroy() {
    this.destoryed$.next();
    this.destoryed$.complete();
  }

  openWizard() {
    this.overlayRef = this.overlay.showModal(this.wizardRef!, {
      hasBackdrop: true,
      backdropStyleOnly: true,
      positionStrategy: new GlobalPositionStrategy().centerHorizontally().centerVertically(),
    });
  }

  closeWizard() {
    this.overlayRef?.close();
  }
}
