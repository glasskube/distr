import {Component, inject, OnInit, TemplateRef, ViewChild, ViewContainerRef} from '@angular/core';
import {DeploymentsComponent} from './deployments.component';
import {InstallationWizardComponent} from '../installation-wizard/installation-wizard.component';
import {EmbeddedOverlayRef, OverlayService} from '../../services/overlay.service';
import {combineLatest, first} from 'rxjs';
import {GlobalPositionStrategy} from '@angular/cdk/overlay';
import {DeploymentTargetsService} from '../../services/deployment-targets.service';
import {ApplicationsService} from '../../services/applications.service';
import {IconDefinition} from '@fortawesome/angular-fontawesome';
import {faPlus} from '@fortawesome/free-solid-svg-icons';

@Component({
  selector: 'app-deployments-page',
  imports: [DeploymentsComponent, InstallationWizardComponent],
  templateUrl: './deployments-page.component.html',
})
export class DeploymentsPageComponent implements OnInit {
  private overlay = inject(OverlayService);

  private readonly deploymentTargets = inject(DeploymentTargetsService);
  deploymentTargets$ = this.deploymentTargets.list();

  private readonly applications = inject(ApplicationsService);
  readonly applications$ = this.applications.list();

  private readonly viewContainerRef = inject(ViewContainerRef);
  @ViewChild('installationWizard') wizardRef?: TemplateRef<unknown>;

  private overlayRef?: EmbeddedOverlayRef;

  protected readonly faPlus: IconDefinition = faPlus;

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
