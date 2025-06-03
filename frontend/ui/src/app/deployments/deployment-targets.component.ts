import {GlobalPositionStrategy, OverlayModule} from '@angular/cdk/overlay';
import {AsyncPipe} from '@angular/common';
import {AfterViewInit, Component, inject, OnDestroy, signal, TemplateRef, ViewChild} from '@angular/core';
import {FormControl, FormGroup, FormsModule, ReactiveFormsModule} from '@angular/forms';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faLightbulb, faMagnifyingGlass, faPlus} from '@fortawesome/free-solid-svg-icons';
import {ApplicationVersion, DeploymentTarget, DeploymentWithLatestRevision} from '@glasskube/distr-sdk';
import {
  catchError,
  combineLatest,
  combineLatestWith,
  first,
  map,
  Observable,
  of,
  Subject,
  switchMap,
  takeUntil,
} from 'rxjs';
import {SemVer} from 'semver';
import {maxBy} from '../../util/arrays';
import {isArchived} from '../../util/dates';
import {filteredByFormControl} from '../../util/filter';
import {drawerFlyInOut} from '../animations/drawer';
import {modalFlyInOut} from '../animations/modal';
import {InstallationWizardComponent} from '../components/installation-wizard/installation-wizard.component';
import {ApplicationsService} from '../services/applications.service';
import {AuthService} from '../services/auth.service';
import {DeploymentTargetsService} from '../services/deployment-targets.service';
import {LicensesService} from '../services/licenses.service';
import {DialogRef, OverlayService} from '../services/overlay.service';
import {DeploymentModalComponent} from './deployment-modal.component';
import {DeploymentTargetCardComponent} from './deployment-target-card/deployment-target-card.component';
import {
  DeploymentTargetLatestMetrics,
  DeploymentTargetsMetricsService,
} from '../services/deployment-target-metrics.service';

type DeploymentWithNewerVersion = {dt: DeploymentTarget; d: DeploymentWithLatestRevision; version: ApplicationVersion};

export interface DeploymentTargetViewData extends DeploymentTarget {
  metrics?: DeploymentTargetLatestMetrics;
}

@Component({
  selector: 'app-deployment-targets',
  imports: [
    AsyncPipe,
    FaIconComponent,
    FormsModule,
    ReactiveFormsModule,
    InstallationWizardComponent,
    OverlayModule,
    DeploymentTargetCardComponent,
    DeploymentModalComponent,
  ],
  templateUrl: './deployment-targets.component.html',
  standalone: true,
  animations: [modalFlyInOut, drawerFlyInOut],
})
export class DeploymentTargetsComponent implements AfterViewInit, OnDestroy {
  public readonly auth = inject(AuthService);
  private readonly overlay = inject(OverlayService);
  private readonly applications = inject(ApplicationsService);
  private readonly licenses = inject(LicensesService);
  private readonly deploymentTargets = inject(DeploymentTargetsService);
  private readonly deploymentTargetMetrics = inject(DeploymentTargetsMetricsService);

  readonly faMagnifyingGlass = faMagnifyingGlass;
  readonly plusIcon = faPlus;
  protected readonly faLightbulb = faLightbulb;

  private destroyed$ = new Subject<void>();
  private modal?: DialogRef;

  @ViewChild('deploymentWizard') protected readonly deploymentWizard!: TemplateRef<unknown>;
  @ViewChild('deploymentModal') protected readonly deploymentModal!: TemplateRef<unknown>;

  selectedDeploymentTarget = signal<DeploymentTarget | undefined>(undefined);
  selectedDeployment = signal<DeploymentWithLatestRevision | undefined>(undefined);
  selectedApplicationVersionId = signal<string | undefined>(undefined);

  readonly filterForm = new FormGroup({
    search: new FormControl(''),
  });

  readonly deploymentTargets$ = this.deploymentTargets.poll().pipe(takeUntil(this.destroyed$));
  readonly deploymentTargetMetrics$ = this.deploymentTargetMetrics.poll().pipe(
    takeUntil(this.destroyed$),
    catchError(() => of([]))
  );

  readonly filteredDeploymentTargets$: Observable<DeploymentTargetViewData[]> = filteredByFormControl(
    this.deploymentTargets$,
    this.filterForm.controls.search,
    (dt, search) => !search || (dt.name || '').toLowerCase().includes(search.toLowerCase())
  ).pipe(
    combineLatestWith(this.deploymentTargetMetrics$),
    map(([deploymentTargets, deploymentTargetMetrics]) => {
      return deploymentTargets.map((dt) => {
        return {
          ...dt,
          metrics: deploymentTargetMetrics.find((x) => x.id === dt.id),
        } as DeploymentTargetViewData;
      });
    })
  );
  private readonly applications$ = this.applications.list();

  protected deploymentTargetsWithUpdate$: Observable<DeploymentWithNewerVersion[]> = this.deploymentTargets$.pipe(
    switchMap((deploymentTargets) =>
      combineLatest(
        deploymentTargets
          .map((dt) =>
            dt.deployments.map((d) =>
              this.getAvailableVersions(d).pipe(
                map((versions) => {
                  const version = this.findMaxVersion(versions);
                  if (version) {
                    return {dt, d, version};
                  }
                  return undefined;
                })
              )
            )
          )
          .flat()
      )
    ),
    map((dts) => dts.filter((dt) => dt !== undefined)),
    map((result) => result.filter((it) => it.version.id !== it.d.applicationVersionId))
  );

  ngAfterViewInit() {
    combineLatest([this.applications$, this.deploymentTargets$])
      .pipe(first())
      .subscribe(([apps, dts]) => {
        if (this.auth.hasRole('customer') && apps.length > 0 && dts.length === 0) {
          this.openWizard();
        }
      });
  }

  ngOnDestroy(): void {
    this.destroyed$.next();
    this.destroyed$.complete();
  }

  protected showDeploymentModal(
    deploymentTarget: DeploymentTarget,
    deployment: DeploymentWithLatestRevision,
    version: ApplicationVersion
  ) {
    this.selectedDeploymentTarget.set(deploymentTarget);
    this.selectedDeployment.set(deployment);
    this.selectedApplicationVersionId.set(version?.id);
    this.hideModal();
    this.modal = this.overlay.showModal(this.deploymentModal, {
      positionStrategy: new GlobalPositionStrategy().centerHorizontally().centerVertically(),
    });
  }

  protected openWizard() {
    this.hideModal();
    this.modal = this.overlay.showModal(this.deploymentWizard, {
      hasBackdrop: true,
      backdropStyleOnly: true,
      positionStrategy: new GlobalPositionStrategy().centerHorizontally().centerVertically(),
    });
  }

  protected hideModal(): void {
    this.modal?.close();
  }

  private getAvailableVersions(deployment: DeploymentWithLatestRevision): Observable<ApplicationVersion[]> {
    return (
      deployment.applicationLicenseId
        ? this.licenses
            .list()
            .pipe(
              map((licenses) => licenses.find((license) => license.id === deployment.applicationLicenseId)?.versions)
            )
        : of(undefined)
    ).pipe(
      switchMap((versions) =>
        versions?.length
          ? of(versions)
          : this.applications$.pipe(
              map((apps) => apps.find((app) => app.id === deployment.applicationId)?.versions ?? [])
            )
      ),
      map((versions) => versions.filter((version) => !isArchived(version)))
    );
  }

  private findMaxVersion(versions: ApplicationVersion[]): ApplicationVersion | undefined {
    try {
      return maxBy(
        versions,
        (version) => new SemVer(version.name),
        (a, b) => a.compare(b) > 0
      );
    } catch (e) {
      console.warn('semver compare failed, falling back to creation date', e);
      return maxBy(versions, (version) => new Date(version.createdAt!));
    }
  }
}
