import {Component, inject, OnInit, TemplateRef, ViewChild, ViewContainerRef} from '@angular/core';
import {GlobeComponent} from '../globe/globe.component';
import {DeploymentTargetsService} from '../../services/deployment-targets.service';
import {ApplicationsComponent} from '../../applications/applications.component';
import {DeploymentTargetsComponent} from '../../deployment-targets/deployment-targets.component';
import {AsyncPipe} from '@angular/common';
import {EmbeddedOverlayRef, OverlayService} from '../../services/overlay.service';
import {OnboardingWizardComponent} from '../onboarding-wizard/onboarding-wizard.component';
import {GlobalPositionStrategy} from '@angular/cdk/overlay';
import {ApplicationsService} from '../../services/applications.service';
import {combineLatest, first} from 'rxjs';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faPlus} from '@fortawesome/free-solid-svg-icons';
import {ChartVersionComponent} from '../charts/version/chart-version.component';
import {ChartUptimeComponent} from '../charts/uptime/chart-uptime.component';
import {ChartTypeComponent} from '../charts/type/chart-type.component';

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
export class DashboardPlaceholderComponent implements OnInit {
  private overlay = inject(OverlayService);
  private readonly deploymentTargets = inject(DeploymentTargetsService);
  public readonly deploymentTargets$ = this.deploymentTargets.list();
  private readonly applications = inject(ApplicationsService);
  readonly applications$ = this.applications.list();

  private viewContainerRef = inject(ViewContainerRef);
  @ViewChild('onboardingWizard') wizardRef?: TemplateRef<unknown>;

  private overlayRef?: EmbeddedOverlayRef;

  protected readonly faPlus = faPlus;

  ngOnInit() {
    const always = false;
    combineLatest([this.applications$, this.deploymentTargets$])
      .pipe(first())
      .subscribe(([apps, dts]) => {
        if (always || apps.length === 0 || dts.length === 0) {
          this.overlayRef?.close();
          this.overlayRef = this.overlay.showModal(this.wizardRef!, this.viewContainerRef, {
            hasBackdrop: true,
            backdropStyleOnly: true,
            positionStrategy: new GlobalPositionStrategy().centerHorizontally().centerVertically(),
          });
        }
      });
  }

  closeWizard() {
    this.overlayRef?.close();
  }
}
