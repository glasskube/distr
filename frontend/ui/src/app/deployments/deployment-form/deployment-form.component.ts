import {AsyncPipe} from '@angular/common';
import {AfterViewInit, Component, forwardRef, inject, Injector, input, OnDestroy, OnInit} from '@angular/core';
import {
  ControlValueAccessor,
  FormBuilder,
  NG_VALUE_ACCESSOR,
  NgControl,
  ReactiveFormsModule,
  TouchedChangeEvent,
  Validators,
} from '@angular/forms';
import {DeploymentRequest} from '@glasskube/distr-sdk';
import {
  catchError,
  combineLatest,
  combineLatestWith,
  debounceTime,
  distinctUntilChanged,
  filter,
  map,
  NEVER,
  of,
  shareReplay,
  Subject,
  switchMap,
  takeUntil,
  withLatestFrom,
} from 'rxjs';
import {isArchived} from '../../../util/dates';
import {HELM_RELEASE_NAME_MAX_LENGTH, HELM_RELEASE_NAME_REGEX} from '../../../util/validation';
import {EditorComponent} from '../../components/editor.component';
import {AutotrimDirective} from '../../directives/autotrim.directive';
import {ApplicationsService} from '../../services/applications.service';
import {FeatureFlagService} from '../../services/feature-flag.service';
import {LicensesService} from '../../services/licenses.service';

export type DeploymentFormValue = Partial<{
  deploymentId: string;
  applicationId: string;
  applicationVersionId: string;
  applicationLicenseId: string;
  valuesYaml: string;
  releaseName: string;
  envFileData: string;
  swarmMode: boolean;
  logsEnabled: boolean;
  forceRestart: boolean;
}>;

export function mapToDeploymentRequest(value: DeploymentFormValue, deploymentTargetId: string): DeploymentRequest {
  return {
    deploymentTargetId: deploymentTargetId,
    applicationVersionId: value.applicationVersionId!,
    applicationLicenseId: value.applicationLicenseId || undefined,
    deploymentId: value.deploymentId || undefined,
    releaseName: value.releaseName || undefined,
    valuesYaml: value.valuesYaml ? btoa(value.valuesYaml) : undefined,
    dockerType: value.swarmMode ? 'swarm' : 'compose',
    envFileData: value.envFileData ? btoa(value.envFileData) : undefined,
    logsEnabled: value.logsEnabled ?? false,
    forceRestart: value.forceRestart ?? false,
  };
}

type DeploymentFormValueCallback = (v: DeploymentFormValue | undefined) => void;

@Component({
  selector: 'app-deployment-form',
  imports: [ReactiveFormsModule, AsyncPipe, EditorComponent, AutotrimDirective],
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => DeploymentFormComponent),
      multi: true,
    },
  ],
  templateUrl: './deployment-form.component.html',
})
export class DeploymentFormComponent implements OnInit, AfterViewInit, OnDestroy, ControlValueAccessor {
  disableApplicationSelect = input(false);
  deploymentType = input<'docker' | 'kubernetes'>('docker');
  customerOrganizationId = input<string>('');
  deploymentTargetName = input<string>('default');

  protected readonly featureFlags = inject(FeatureFlagService);
  private readonly applications = inject(ApplicationsService);
  private readonly licenses = inject(LicensesService);
  private readonly fb = inject(FormBuilder);
  private readonly injector = inject(Injector);

  protected readonly deployForm = this.fb.nonNullable.group({
    deploymentId: this.fb.nonNullable.control<string | undefined>(undefined),
    applicationId: this.fb.nonNullable.control('', Validators.required),
    applicationVersionId: this.fb.nonNullable.control('', Validators.required),
    applicationLicenseId: this.fb.nonNullable.control('', Validators.required),
    releaseName: this.fb.nonNullable.control('', [
      Validators.required,
      Validators.maxLength(HELM_RELEASE_NAME_MAX_LENGTH),
      Validators.pattern(HELM_RELEASE_NAME_REGEX),
    ]),
    valuesYaml: this.fb.nonNullable.control(''),
    envFileData: this.fb.nonNullable.control(''),
    swarmMode: this.fb.nonNullable.control<boolean>(false),
    logsEnabled: this.fb.nonNullable.control<boolean>(true),
    forceRestart: this.fb.nonNullable.control<boolean>(false),
  });
  protected readonly composeFile = this.fb.nonNullable.control({disabled: true, value: ''});

  private readonly deploymentId$ = this.deployForm.controls.deploymentId.valueChanges.pipe(
    distinctUntilChanged(),
    shareReplay(1)
  );

  private readonly applicationId$ = this.deployForm.controls.applicationId.valueChanges.pipe(
    distinctUntilChanged(),
    shareReplay(1)
  );

  private readonly applicationVersionId$ = this.deployForm.controls.applicationVersionId.valueChanges.pipe(
    distinctUntilChanged(),
    shareReplay(1)
  );

  private readonly applicationLicenseId$ = this.deployForm.controls.applicationLicenseId.valueChanges.pipe(
    distinctUntilChanged(),
    shareReplay(1)
  );

  /**
   * The license control is VISIBLE for users editing a customer managed deployment.
   */
  protected readonly licenseControlVisible$ = combineLatest([
    this.featureFlags.isLicensingEnabled$,
    this.licenses.list(),
  ]).pipe(
    map(
      ([isLicensingEnabled, licenses]) =>
        isLicensingEnabled && this.customerOrganizationId() !== '' && licenses.length > 0
    ),
    distinctUntilChanged()
  );

  /**
   * The license control is ENABLED when deploying to a customer managed target and there is no deployment yet.
   * A vendor might be required to choose a license for a customer managed deployment target with no previous
   * deployment but they may only choose a license owned by the same customer.
   */
  private readonly licenseControlEnabled$ = combineLatest([this.licenseControlVisible$, this.deploymentId$]).pipe(
    map(([isVisible, deploymentId]) => isVisible && !deploymentId),
    distinctUntilChanged()
  );

  protected readonly swarmModeVisible$ = of(this.deploymentType() === 'docker');

  protected readonly applications$ = this.applications
    .list()
    .pipe(map((apps) => apps.filter((app) => app.type === this.deploymentType())));

  private selectedApplication$ = this.applicationId$.pipe(
    combineLatestWith(this.applications$),
    map(([applicationId, applications]) => applications.find((application) => application.id === applicationId))
  );

  protected readonly licenses$ = this.applicationId$.pipe(
    combineLatestWith(this.licenseControlVisible$),
    switchMap(([applicationId, isLicensingEnabled]) =>
      isLicensingEnabled && applicationId ? this.licenses.list(applicationId) : NEVER
    ),
    map((licenses) => licenses.filter((l) => l.customerOrganizationId === this.customerOrganizationId())),
    shareReplay(1)
  );

  private readonly selectedLicense$ = this.applicationLicenseId$.pipe(
    combineLatestWith(this.licenses$),
    map(([licenseId, licenses]) => licenses.find((license) => license.id === licenseId))
  );

  protected availableApplicationVersions$ = this.licenseControlVisible$.pipe(
    switchMap((shouldShowLicense) =>
      shouldShowLicense
        ? this.selectedLicense$.pipe(
            switchMap((license) =>
              // if the license has no version associations, assume that the application has all available versions
              license?.versions?.length
                ? of(license.versions)
                : this.selectedApplication$.pipe(map((application) => application?.versions ?? []))
            )
          )
        : this.selectedApplication$.pipe(map((application) => application?.versions ?? []))
    ),
    withLatestFrom(this.applicationVersionId$),
    map(([avs, selectedApplicationVersionId]) =>
      avs.filter((av) => {
        if (av.id === selectedApplicationVersionId) {
          return true;
        }
        return !isArchived(av);
      })
    )
  );

  private readonly destroyed$ = new Subject<void>();

  private onChange?: DeploymentFormValueCallback;
  private onTouched?: DeploymentFormValueCallback;

  ngOnInit(): void {
    combineLatest([this.deployForm.valueChanges, this.deployForm.statusChanges])
      .pipe(takeUntil(this.destroyed$))
      .subscribe(([value, status]) => {
        const callbackArg = status === 'VALID' ? value : undefined;
        this.onChange?.(callbackArg);
        this.onTouched?.(callbackArg);
      });

    this.licenseControlEnabled$.pipe(takeUntil(this.destroyed$)).subscribe((licenseControlEnabled) => {
      if (licenseControlEnabled) {
        this.deployForm.controls.applicationLicenseId.enable();
      } else {
        this.deployForm.controls.applicationLicenseId.disable();
      }
    });

    const type = this.deploymentType();
    if (type) {
      if (type === 'kubernetes') {
        this.deployForm.controls.releaseName.enable();
        this.deployForm.controls.valuesYaml.enable();
        this.deployForm.controls.envFileData.disable();
        const targetName = this.deploymentTargetName();
        if (!this.deployForm.value.releaseName && targetName) {
          this.deployForm.patchValue({
            releaseName: targetName.trim().toLowerCase().replaceAll(/\W+/g, '-'),
          });
        }
      } else {
        this.deployForm.controls.envFileData.enable();
        this.deployForm.controls.releaseName.disable();
        this.deployForm.controls.valuesYaml.disable();
      }
    }

    this.deploymentId$.pipe(takeUntil(this.destroyed$)).subscribe((id) => {
      if (id) {
        this.deployForm.controls.applicationId.disable();
        this.deployForm.controls.swarmMode.disable();
        this.deployForm.controls.forceRestart.enable();
      } else {
        if (!this.disableApplicationSelect()) {
          this.deployForm.controls.applicationId.enable();
        }
        this.deployForm.controls.swarmMode.enable();
        this.deployForm.controls.forceRestart.disable();
      }
    });

    combineLatest([this.applicationId$, this.applicationVersionId$, this.deploymentId$])
      .pipe(
        debounceTime(5),
        switchMap(([applicationId, versionId, deploymentId]) =>
          versionId && applicationId && !deploymentId
            ? this.applications.getTemplateFile(applicationId, versionId).pipe(catchError(() => NEVER))
            : NEVER
        ),
        takeUntil(this.destroyed$)
      )
      .subscribe((templateFile) => {
        const type = this.deploymentType();
        if (type === 'kubernetes') {
          this.deployForm.controls.valuesYaml.patchValue(templateFile ?? '');
        } else {
          this.deployForm.controls.envFileData.patchValue(templateFile ?? '');
        }
      });

    combineLatest([this.applicationId$, this.applicationVersionId$])
      .pipe(
        debounceTime(5),
        switchMap(([applicationId, versionId]) =>
          versionId && applicationId && this.deploymentType() === 'docker'
            ? this.applications.getComposeFile(applicationId, versionId).pipe(catchError(() => NEVER))
            : NEVER
        ),
        takeUntil(this.destroyed$)
      )
      .subscribe((composeFile) => {
        this.composeFile.patchValue(composeFile ?? '');
      });

    this.licenses$.pipe(takeUntil(this.destroyed$)).subscribe((licenses) => {
      // Only update the form control, if the previously selected license is no longer in the list
      if (
        licenses.length > 0 &&
        licenses[0].id &&
        licenses.every((l) => l.id !== this.deployForm.controls.applicationLicenseId.value)
      ) {
        this.deployForm.controls.applicationLicenseId.setValue(licenses[0].id);
      }
    });

    this.availableApplicationVersions$.pipe(takeUntil(this.destroyed$)).subscribe((versions) => {
      if (versions.length > 0) {
        this.deployForm.controls.applicationVersionId.enable();
        const version = versions[versions.length - 1];
        // Only update the form control, if the previously selected version is no longer in the list
        if (version.id && versions.every((version) => version.id !== this.deployForm.value.applicationVersionId)) {
          this.deployForm.controls.applicationVersionId.setValue(version.id);
        }
      } else {
        this.deployForm.controls.applicationVersionId.disable();
        // this.deployForm.controls.applicationVersionId.reset();
      }
    });

    // This is needed because the first value could be missed otherwise
    // TODO: Find a better solution for this
    this.applicationLicenseId$.pipe(takeUntil(this.destroyed$)).subscribe();

    // Disable application selector if requested
    if (this.disableApplicationSelect()) {
      this.deployForm.controls.applicationId.disable();
    }
  }

  ngAfterViewInit(): void {
    // adapted from https://github.com/angular/angular/issues/45089
    this.injector
      .get(NgControl)
      .control!.events.pipe(
        takeUntil(this.destroyed$),
        filter((event) => event instanceof TouchedChangeEvent && event.touched)
      )
      .subscribe(() => this.deployForm.markAllAsTouched());
  }

  ngOnDestroy(): void {
    this.destroyed$.next();
    this.destroyed$.complete();
  }

  registerOnChange(fn: DeploymentFormValueCallback): void {
    this.onChange = fn;
  }

  registerOnTouched(fn: DeploymentFormValueCallback): void {
    this.onTouched = fn;
  }

  setDisabledState(isDisabled: boolean): void {
    if (isDisabled) {
      console.warn('DeploymentFormComponent does not support setDisabledState');
    }
  }

  writeValue(obj: DeploymentFormValue | null | undefined): void {
    if (obj) {
      this.deployForm.patchValue(obj);
    } else {
      this.deployForm.reset();
    }
  }
}
