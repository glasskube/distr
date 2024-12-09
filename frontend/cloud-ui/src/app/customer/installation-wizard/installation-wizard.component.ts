import {Component, ElementRef, EventEmitter, inject, Input, Output, TemplateRef, ViewChild} from '@angular/core';
import {DeploymentTargetsService} from '../../services/deployment-targets.service';
import {faCopy, faShip, faXmark} from '@fortawesome/free-solid-svg-icons';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {modalFlyInOut} from '../../animations/modal';
import {CdkStep, CdkStepper} from '@angular/cdk/stepper';
import {ApplicationsService} from '../../services/applications.service';
import {firstValueFrom, Observable, Subject, switchMap, takeUntil, tap, withLatestFrom} from 'rxjs';
import {DeploymentService} from '../../services/deployment.service';
import {Application} from '../../types/application';
import {DeploymentTarget} from '../../types/deployment-target';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {InstallationWizardStepperComponent} from './installation-wizard-stepper.component';
import {AsyncPipe} from '@angular/common';
import {Deployment} from '../../types/deployment';
import {DeploymentTargetViewModel} from '../../deployment-targets/DeploymentTargetViewModel';

@Component({
  selector: 'app-installation-wizard',
  templateUrl: './installation-wizard.component.html',
  imports: [ReactiveFormsModule, FaIconComponent, InstallationWizardStepperComponent, CdkStep, AsyncPipe],
  animations: [modalFlyInOut],
})
export class InstallationWizardComponent {
  protected readonly xmarkIcon = faXmark;
  private applications = inject(ApplicationsService);
  private deploymentTargets = inject(DeploymentTargetsService);
  private deployments = inject(DeploymentService);
  @ViewChild('stepper') stepper!: CdkStepper;

  createdApp?: Application;
  createdDeploymentTarget?: DeploymentTarget;

  @Output('closed') closed = new EventEmitter<void>();

  applicationForm = new FormGroup({
    type: new FormControl<string>('sample', Validators.required),
    custom: new FormGroup({
      name: new FormControl<string>(
        {
          value: '',
          disabled: true,
        },
        Validators.required
      ),
      versionName: new FormControl<string>(
        {
          value: '',
          disabled: true,
        },
        Validators.required
      ),
    }),
  });
  fileToUpload: File | null = null;
  @ViewChild('fileInput')
  fileInput?: ElementRef;

  deploymentTargetForm = new FormGroup({});

  private loading = false;

  ngOnInit() {
    this.applicationForm.controls.type.valueChanges.pipe(takeUntil(this.destroyed$)).subscribe((type) => {
      if (type === 'sample') {
        // disable validators
        this.applicationForm.controls.custom.controls.name.disable();
        this.applicationForm.controls.custom.controls.versionName.disable();
      } else {
        this.applicationForm.controls.custom.controls.name.enable();
        this.applicationForm.controls.custom.controls.versionName.enable();
      }
    });
  }

  private destroyed$: Subject<void> = new Subject();

  ngOnDestroy() {
    this.destroyed$.complete();
    this.applicationIdChange$.unsubscribe();
  }

  onFileSelected(event: Event) {
    this.fileToUpload = (event.target as HTMLInputElement).files?.[0] ?? null;
  }

  attemptContinue() {
    if (this.loading) {
      return;
    }

    if (this.stepper.selectedIndex === 0) {
      if (this.applicationForm.valid) {
        if (this.applicationForm.controls.type.value === 'sample') {
          this.loading = true;
          this.applications.createSample().subscribe((app) => {
            this.createdApp = app;
            this.nextStep();
          });
        } else if (this.fileToUpload != null) {
          this.loading = true;
          this.applications
            .create({
              name: this.applicationForm.controls.custom.controls.name.value!,
              type: 'docker',
            })
            .pipe(
              switchMap((application) =>
                this.applications.createApplicationVersion(
                  application,
                  {
                    name: this.applicationForm.controls.custom.controls.versionName.value!,
                  },
                  this.fileToUpload!
                )
              ),
              withLatestFrom(this.applications.list())
            )
            .subscribe(([version, apps]) => {
              this.createdApp = apps.find((a) => a.id === version.applicationId);
              this.nextStep();
            });
        }
      }
    } else if (this.stepper.selectedIndex === 1) {
      if (this.deploymentTargetForm.valid) {
        this.loading = true;
        this.nextStep();
      }
    } else if (this.stepper.selectedIndex == 2) {
      this.close();
    }
  }

  close() {
    this.closed.emit();
  }

  private nextStep() {
    this.loading = false;
    this.stepper.next();
  }

  protected readonly shipIcon = faShip;

  @Input({required: true})
  applications$!: Observable<Application[]>;
  selectedApplication?: Application | null;

  @Input({required: true})
  deploymentTargets$!: Observable<DeploymentTargetViewModel[]>;

  deployForm = new FormGroup({
    deploymentTargetId: new FormControl<string | undefined>(undefined, Validators.required),
    applicationId: new FormControl<string | undefined>(undefined, Validators.required),
    applicationVersionId: new FormControl<string | undefined>(undefined, Validators.required),
    notes: new FormControl<string | undefined>(undefined),
  });

  private selectedDeploymentTarget?: DeploymentTargetViewModel | null;

  private readonly applicationIdChange$ = this.deployForm.controls.applicationId.valueChanges.subscribe((it) =>
    this.updatedSelectedApplication(it!!)
  );

  async newDeployment(dt: DeploymentTarget, deploymentModal: TemplateRef<any>) {
    this.selectedDeploymentTarget = dt;
    this.deployForm.patchValue({
      deploymentTargetId: dt.id,
    });
    this.deploymentTargets.latestDeploymentFor(dt.id!!).subscribe((d) => {
      this.deployForm.patchValue({
        applicationId: d.applicationId,
        applicationVersionId: d.applicationVersionId,
      });
      this.updatedSelectedApplication(d.applicationId);
    });
  }

  async saveDeployment() {
    if (this.deployForm.valid) {
      const deployment = this.deployForm.value;
      await firstValueFrom(this.deployments.create(deployment as Deployment));
      this.selectedDeploymentTarget!!.latestDeployment = this.deploymentTargets.latestDeploymentFor(
        this.selectedDeploymentTarget!!.id!!
      );
    }
  }

  async updatedSelectedApplication(applicationId: string) {
    let applications = await firstValueFrom(this.applications$);
    this.selectedApplication = applications.find((a) => a.id === applicationId) || null;
  }
}
