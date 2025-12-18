import {CdkStep, CdkStepper} from '@angular/cdk/stepper';
import {AsyncPipe} from '@angular/common';
import {Component, computed, EventEmitter, inject, OnDestroy, OnInit, Output, signal, ViewChild} from '@angular/core';
import {toSignal} from '@angular/core/rxjs-interop';
import {FormControl, FormGroup, FormsModule, ReactiveFormsModule, Validators} from '@angular/forms';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faDocker} from '@fortawesome/free-brands-svg-icons';
import {faBuildingUser, faCheckCircle, faDharmachakra, faShip, faXmark} from '@fortawesome/free-solid-svg-icons';
import {
  Application,
  CustomerOrganization,
  DeploymentTarget,
  DeploymentTargetScope,
  DeploymentType,
} from '@glasskube/distr-sdk';
import {combineLatest, firstValueFrom, map, startWith, Subject, takeUntil} from 'rxjs';
import {getFormDisplayedError} from '../../../util/errors';
import {SecureImagePipe} from '../../../util/secureImage';
import {KUBERNETES_RESOURCE_MAX_LENGTH, KUBERNETES_RESOURCE_NAME_REGEX} from '../../../util/validation';
import {modalFlyInOut} from '../../animations/modal';
import {ConnectInstructionsComponent} from '../../components/connect-instructions/connect-instructions.component';
import {AutotrimDirective} from '../../directives/autotrim.directive';
import {ApplicationsService} from '../../services/applications.service';
import {AuthService} from '../../services/auth.service';
import {CustomerOrganizationsService} from '../../services/customer-organizations.service';
import {DeploymentTargetsService} from '../../services/deployment-targets.service';
import {FeatureFlagService} from '../../services/feature-flag.service';
import {LicensesService} from '../../services/licenses.service';
import {OrganizationBrandingService} from '../../services/organization-branding.service';
import {OrganizationService} from '../../services/organization.service';
import {ToastService} from '../../services/toast.service';
import {DeploymentFormComponent, mapToDeploymentRequest} from '../deployment-form/deployment-form.component';
import {DeploymentWizardStepperComponent} from './deployment-wizard-stepper.component';

@Component({
  selector: 'app-deployment-wizard',
  templateUrl: './deployment-wizard.component.html',
  imports: [
    AsyncPipe,
    ReactiveFormsModule,
    FormsModule,
    FaIconComponent,
    DeploymentWizardStepperComponent,
    CdkStep,
    ConnectInstructionsComponent,
    AutotrimDirective,
    SecureImagePipe,
    DeploymentFormComponent,
  ],
  animations: [modalFlyInOut],
})
export class DeploymentWizardComponent implements OnInit, OnDestroy {
  protected readonly xmarkIcon = faXmark;
  protected readonly shipIcon = faShip;
  protected readonly dockerIcon = faDocker;
  protected readonly kubernetesIcon = faDharmachakra;
  protected readonly buildingUserIcon = faBuildingUser;
  protected readonly checkCircleIcon = faCheckCircle;

  private readonly toast = inject(ToastService);
  private readonly applications = inject(ApplicationsService);
  private readonly deploymentTargets = inject(DeploymentTargetsService);
  private readonly customerOrganizations = inject(CustomerOrganizationsService);
  private readonly licenses = inject(LicensesService);
  private readonly organization = inject(OrganizationService);
  private readonly organizationBranding = inject(OrganizationBrandingService);
  protected readonly auth = inject(AuthService);
  protected readonly featureFlags = inject(FeatureFlagService);

  @ViewChild('stepper') private stepper?: CdkStepper;

  @Output('closed') readonly closed = new EventEmitter<void>();

  // Step 1: Customer Selection (optional)
  readonly customerForm = new FormGroup({
    customerOrganizationId: new FormControl<string | null>(null),
  });

  // Step 2: Application Selection
  readonly applicationForm = new FormGroup({
    applicationId: new FormControl<string>('', Validators.required),
  });

  // Step 3: Deployment Target Configuration
  readonly deploymentTargetForm = new FormGroup({
    name: new FormControl<string>('', Validators.required),
    namespace: new FormControl<string>('', [
      Validators.required,
      Validators.maxLength(KUBERNETES_RESOURCE_MAX_LENGTH),
      Validators.pattern(KUBERNETES_RESOURCE_NAME_REGEX),
    ]),
    clusterScope: new FormControl<boolean>(true, {nonNullable: true}),
    scope: new FormControl<DeploymentTargetScope>('cluster', {nonNullable: true}),
  });

  // Step 4: Application Configuration
  readonly applicationConfigForm = new FormGroup({
    deploymentFormData: new FormControl<any>(null, Validators.required),
  });

  // Step 5: Connect Instructions (no form, just display)
  readonly connectForm = new FormGroup({});

  // State management
  protected readonly customerOrganizations$ = this.customerOrganizations.getCustomerOrganizations();
  protected readonly applications$ = this.applications.list();
  protected readonly allLicenses$ = this.licenses.list();
  protected readonly vendorOrganization$ = this.organization.get();
  protected readonly vendorBranding$ = this.organizationBranding.get();
  protected selectedApplication = signal<Application | undefined>(undefined);
  protected selectedCustomerOrganization = signal<CustomerOrganization | undefined>(undefined);
  protected selectedDeploymentTarget = signal<DeploymentTarget | null>(null);

  // Filter applications based on customer licenses
  protected readonly filteredApplications$ = combineLatest([
    this.applications$,
    this.allLicenses$,
    this.customerForm.controls.customerOrganizationId.valueChanges.pipe(
      startWith(this.customerForm.controls.customerOrganizationId.value),
      map((id) => id ?? undefined)
    ),
    this.featureFlags.isLicensingEnabled$,
  ]).pipe(
    map(([applications, licenses, customerOrgId, isLicensingEnabled]) => {
      // If licensing is not enabled or no customer is selected, show all applications
      if (!isLicensingEnabled || !customerOrgId) {
        return applications;
      }

      // Filter applications to only show those with licenses for the selected customer
      const customerLicenses = licenses.filter((l) => l.customerOrganizationId === customerOrgId);
      const licensedApplicationIds = new Set(customerLicenses.map((l) => l.applicationId));

      return applications.filter((app) => licensedApplicationIds.has(app.id));
    })
  );

  // Computed properties
  protected readonly showCustomerStep = computed(() => {
    return this.auth.isVendor();
  });

  protected readonly selectedDeploymentType = computed<DeploymentType | undefined>(() => {
    const app = this.selectedApplication();
    return app?.type;
  });

  // Initial data for deployment form
  protected readonly deploymentFormInitialData = computed(() => {
    const app = this.selectedApplication();
    const deploymentTarget = this.selectedDeploymentTarget();
    if (!app || !deploymentTarget) {
      return null;
    }
    return {
      deploymentTargetId: deploymentTarget.id!,
      applicationId: app.id!,
    };
  });

  private readonly isLicensingEnabled = toSignal(this.featureFlags.isLicensingEnabled$, {initialValue: false});

  protected readonly showLicenseControl = computed(() => {
    return this.selectedCustomerOrganization() !== undefined && this.isLicensingEnabled();
  });

  protected getVendorLogoUrl(branding: {logo?: string; logoContentType?: string} | null): string {
    if (branding?.logo && branding?.logoContentType) {
      return `data:${branding.logoContentType};base64,${branding.logo}`;
    }
    return '/distr-logo.svg';
  }

  private loading = false;
  private readonly destroyed$ = new Subject<void>();

  ngOnInit() {
    // Watch customer selection
    this.customerForm.controls.customerOrganizationId.valueChanges
      .pipe(takeUntil(this.destroyed$))
      .subscribe((customerId) => {
        if (customerId) {
          firstValueFrom(this.customerOrganizations$).then((customers) => {
            const customer = customers.find((c) => c.id === customerId);
            this.selectedCustomerOrganization.set(customer);
          });
        } else {
          this.selectedCustomerOrganization.set(undefined);
        }
      });

    // Watch application selection
    this.applicationForm.controls.applicationId.valueChanges.pipe(takeUntil(this.destroyed$)).subscribe((appId) => {
      firstValueFrom(this.applications$).then((apps) => {
        const app = apps.find((a) => a.id === appId);
        this.selectedApplication.set(app);

        // Enable/disable configuration form controls based on deployment type
        this.updateConfigurationFormControls(app?.type);
      });
    });

    // Watch cluster scope changes
    this.deploymentTargetForm.controls.clusterScope.valueChanges.pipe(takeUntil(this.destroyed$)).subscribe((value) => {
      this.deploymentTargetForm.controls.scope.setValue(value ? 'cluster' : 'namespace');
    });
  }

  ngOnDestroy() {
    this.destroyed$.next();
    this.destroyed$.complete();
  }

  private updateConfigurationFormControls(type: DeploymentType | undefined) {
    if (type === 'kubernetes') {
      this.deploymentTargetForm.controls.namespace.enable();
      this.deploymentTargetForm.controls.clusterScope.enable();
      this.deploymentTargetForm.controls.scope.enable();
    } else if (type === 'docker') {
      this.deploymentTargetForm.controls.namespace.disable();
      this.deploymentTargetForm.controls.clusterScope.disable();
      this.deploymentTargetForm.controls.scope.disable();
    }
  }

  async attemptContinue() {
    if (this.loading) {
      return;
    }

    const stepIndex = this.stepper?.selectedIndex ?? 0;
    const adjustedIndex = this.showCustomerStep() ? stepIndex : stepIndex + 1;

    switch (adjustedIndex) {
      case 0:
        // Step 1: Customer Selection
        this.nextStep();
        break;
      case 1:
        // Step 2: Application Selection
        await this.continueFromApplicationStep();
        break;
      case 2:
        // Step 3: Deployment Target Configuration
        await this.continueFromDeploymentTargetStep();
        break;
      case 3:
        // Step 4: Application Configuration
        await this.continueFromApplicationConfigStep();
        break;
      case 4:
        // Step 5: Connect and Deploy
        await this.continueFromConnectStep();
        break;
    }
  }

  private async continueFromApplicationStep() {
    this.applicationForm.markAllAsTouched();
    if (!this.applicationForm.valid) {
      return;
    }
    this.nextStep();
  }

  private async continueFromDeploymentTargetStep() {
    this.deploymentTargetForm.markAllAsTouched();
    if (!this.deploymentTargetForm.valid || this.loading) {
      return;
    }

    this.loading = true;
    try {
      const app = this.selectedApplication();
      if (!app) {
        throw new Error('No application selected');
      }

      const customerOrgId = this.selectedCustomerOrganization()?.id;

      console.log(customerOrgId);

      // Create deployment target
      const created = await firstValueFrom(
        this.deploymentTargets.create({
          name: this.deploymentTargetForm.value.name!,
          type: app.type!,
          namespace: this.deploymentTargetForm.value.namespace || undefined,
          scope: this.deploymentTargetForm.value.scope,
          deployments: [],
          metricsEnabled: this.deploymentTargetForm.value.scope !== 'namespace',
          customerOrganization: customerOrgId ? ({id: customerOrgId} as CustomerOrganization) : undefined,
        })
      );

      this.selectedDeploymentTarget.set(created as DeploymentTarget);
      this.nextStep();
    } catch (e) {
      const msg = getFormDisplayedError(e);
      if (msg) {
        this.toast.error(msg);
      }
    } finally {
      this.loading = false;
    }
  }

  private async continueFromApplicationConfigStep() {
    this.applicationConfigForm.markAllAsTouched();
    if (!this.applicationConfigForm.valid) {
      return;
    }
    this.nextStep();
  }

  private async continueFromConnectStep() {
    if (this.loading) {
      return;
    }

    try {
      this.loading = true;
      const deploymentFormData = this.applicationConfigForm.value.deploymentFormData;

      if (!deploymentFormData) {
        throw new Error('Missing deployment configuration');
      }

      const deployment = mapToDeploymentRequest(deploymentFormData);

      await firstValueFrom(this.deploymentTargets.deploy(deployment));
      this.toast.success('Deployment created successfully');
      this.close();
    } catch (e) {
      const msg = getFormDisplayedError(e);
      if (msg) {
        this.toast.error(msg);
      }
    } finally {
      this.loading = false;
    }
  }

  close() {
    this.closed.emit();
  }

  private nextStep() {
    this.loading = false;
    this.stepper?.next();
  }

  selectApplication(app: Application) {
    this.applicationForm.controls.applicationId.setValue(app.id!);
  }

  selectCustomer(customer: CustomerOrganization | null) {
    this.customerForm.controls.customerOrganizationId.setValue(customer?.id ?? null);
  }
}
