import {GlobalPositionStrategy} from '@angular/cdk/overlay';
import {AsyncPipe} from '@angular/common';
import {AfterViewInit, Component, inject, OnInit, TemplateRef, ViewChild} from '@angular/core';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faPlus} from '@fortawesome/free-solid-svg-icons';
import {combineLatest, first} from 'rxjs';
import {ApplicationsComponent} from '../../applications/applications.component';
import {DeploymentTargetsComponent} from '../../deployments/deployment-targets.component';
import {ApplicationsService} from '../../services/applications.service';
import {AuthService} from '../../services/auth.service';
import {DeploymentTargetsService} from '../../services/deployment-targets.service';
import {DialogRef, OverlayService} from '../../services/overlay.service';
import {ChartTypeComponent} from '../charts/type/chart-type.component';
import {ChartUptimeComponent} from '../charts/uptime/chart-uptime.component';
import {ChartVersionComponent} from '../charts/version/chart-version.component';
import {GlobeComponent} from '../globe/globe.component';
import {OnboardingWizardComponent} from '../onboarding-wizard/onboarding-wizard.component';
import {UsersService} from '../../services/users.service';

@Component({
  selector: 'app-dashboard-placeholder',
  imports: [
    ApplicationsComponent,
    DeploymentTargetsComponent,
    GlobeComponent,
    AsyncPipe,
    OnboardingWizardComponent,
    FaIconComponent,
    ChartVersionComponent,
    ChartVersionComponent,
    ChartUptimeComponent,
    ChartTypeComponent,
  ],
  templateUrl: './dashboard-placeholder.component.html',
})
export class DashboardPlaceholderComponent implements AfterViewInit {
  private overlay = inject(OverlayService);
  private readonly deploymentTargets = inject(DeploymentTargetsService);
  public readonly deploymentTargets$ = this.deploymentTargets.list();
  private readonly applications = inject(ApplicationsService);
  readonly applications$ = this.applications.list();
  private readonly user = inject(UsersService);
  public readonly users$ = this.user.getUsers();
  private readonly auth = inject(AuthService);

  @ViewChild('onboardingWizard') wizardRef?: TemplateRef<unknown>;

  private overlayRef?: DialogRef;

  protected readonly faPlus = faPlus;

  ngAfterViewInit() {
    combineLatest([this.applications$])
      .pipe(first())
      .subscribe(([apps]) => {
        if (true || (this.auth.hasRole('vendor') && apps.length === 0)) {
          this.closeWizard();
          this.openWizard();
        }
      });
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
