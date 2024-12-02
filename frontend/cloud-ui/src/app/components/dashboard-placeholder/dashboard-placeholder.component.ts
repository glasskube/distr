import {Component, inject, TemplateRef, ViewChild, ViewContainerRef} from '@angular/core';
import {GlobeComponent} from '../globe/globe.component';
import {DeploymentTargetsService} from '../../services/deployment-targets.service';
import {ApplicationsComponent} from '../../applications/applications.component';
import {DeploymentTargetsComponent} from '../../deployment-targets/deployment-targets.component';
import {AsyncPipe} from '@angular/common';
import {OverlayService} from '../../services/overlay.service';
import {OnboardingWizardComponent} from '../onboarding-wizard/onboarding-wizard.component';
import {GlobalPositionStrategy} from '@angular/cdk/overlay';

@Component({
  selector: 'app-dashboard-placeholder',
  imports: [ApplicationsComponent, DeploymentTargetsComponent, GlobeComponent, AsyncPipe, OnboardingWizardComponent],
  templateUrl: './dashboard-placeholder.component.html',
})
export class DashboardPlaceholderComponent {
  private overlay = inject(OverlayService);
  private readonly deploymentTargets = inject(DeploymentTargetsService);
  public readonly deploymentTargets$ = this.deploymentTargets.list();

  private viewContainerRef = inject(ViewContainerRef);
  @ViewChild('onboardingWizard') wizardRef?: TemplateRef<unknown>;

  ngAfterViewInit() {
    this.overlay.showModal(this.wizardRef!, this.viewContainerRef, {
      hasBackdrop: true, // TODO
      positionStrategy: new GlobalPositionStrategy().centerHorizontally().centerVertically()
    });
  }
}
