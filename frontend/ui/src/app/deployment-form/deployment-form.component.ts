import {AsyncPipe} from '@angular/common';
import {Component, forwardRef, inject, OnDestroy, OnInit} from '@angular/core';
import {
  ControlValueAccessor,
  FormBuilder,
  FormGroup,
  NG_VALUE_ACCESSOR,
  ReactiveFormsModule,
  Validators,
} from '@angular/forms';
import {
  catchError,
  combineLatest,
  combineLatestWith,
  debounceTime,
  distinctUntilChanged,
  map,
  NEVER,
  of,
  shareReplay,
  Subject,
  switchMap,
  takeUntil,
} from 'rxjs';
import {YamlEditorComponent} from '../components/yaml-editor.component';
import {AutotrimDirective} from '../directives/autotrim.directive';
import {ApplicationsService} from '../services/applications.service';
import {AuthService} from '../services/auth.service';
import {DeploymentTargetsService} from '../services/deployment-targets.service';
import {FeatureFlagService} from '../services/feature-flag.service';
import {LicensesService} from '../services/licenses.service';

export type DeploymentFormValue = Partial<{
  deploymentTargetId: string;
  applicationId: string;
  applicationVersionId: string;
  applicationLicenseId: string;
  valuesYaml: string;
  releaseName: string;
  envFileData: string;
}>;

type DeploymentFormValueCallback = (v: DeploymentFormValue | undefined) => void;

@Component({
  selector: 'app-deployment-form',
  imports: [ReactiveFormsModule, AsyncPipe, YamlEditorComponent, AutotrimDirective],
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => DeploymentFormComponent),
      multi: true,
    },
  ],
  templateUrl: './deployment-form.component.html',
})
export class DeploymentFormComponent implements OnInit, OnDestroy, ControlValueAccessor {
  protected readonly featureFlags = inject(FeatureFlagService);
  private readonly applications = inject(ApplicationsService);
  private readonly licenses = inject(LicensesService);
  private readonly auth = inject(AuthService);
  private readonly fb = inject(FormBuilder);
  private readonly deplyomentTargets = inject(DeploymentTargetsService);

  protected readonly deployForm = this.fb.nonNullable.group({
    deploymentTargetId: this.fb.nonNullable.control('', Validators.required),
    applicationId: this.fb.nonNullable.control('', Validators.required),
    applicationVersionId: this.fb.nonNullable.control('', Validators.required),
    applicationLicenseId: this.fb.nonNullable.control('', Validators.required),
    releaseName: this.fb.nonNullable.control('', Validators.required),
    valuesYaml: this.fb.nonNullable.control(''),
    envFileData: this.fb.nonNullable.control(''),
  });

  private readonly deploymentTargetId$ = this.deployForm.controls.deploymentTargetId.valueChanges.pipe(
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
  protected readonly shouldShowLicense$ = this.featureFlags.isLicensingEnabled$.pipe(
    map((isLicensingEnabled) => isLicensingEnabled && this.auth.hasRole('customer'))
  );
  private readonly deploymentTarget$ = this.deploymentTargetId$.pipe(
    combineLatestWith(this.deplyomentTargets.list()),
    map(([id, dts]) => dts.find((dt) => dt.id === id)),
    shareReplay(1)
  );
  protected readonly applications$ = this.deploymentTarget$.pipe(
    map((dt) => dt?.type),
    combineLatestWith(this.applications.list()),
    map(([type, apps]) => apps.filter((app) => app.type === type))
  );
  private selectedApplication$ = this.applicationId$.pipe(
    combineLatestWith(this.applications$),
    map(([applicationId, applications]) => applications.find((application) => application.id === applicationId))
  );
  protected readonly licenses$ = this.applicationId$.pipe(
    combineLatestWith(this.shouldShowLicense$),
    switchMap(([applicationId, isLicensingEnabled]) =>
      isLicensingEnabled && applicationId ? this.licenses.list(applicationId) : NEVER
    ),
    shareReplay(1)
  );
  private readonly selectedLicense$ = this.applicationLicenseId$.pipe(
    combineLatestWith(this.licenses$),
    map(([licenseId, licenses]) => licenses.find((license) => license.id === licenseId))
  );
  protected availableApplicationVersions$ = this.shouldShowLicense$.pipe(
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

    combineLatest([
      this.shouldShowLicense$,
      this.deploymentTarget$.pipe(
        map((dt) => !!dt?.deployment),
        distinctUntilChanged()
      ),
    ])
      .pipe(takeUntil(this.destroyed$))
      .subscribe(([isLicensingEnabled, existingDeployment]) => {
        if (isLicensingEnabled && !existingDeployment) {
          this.deployForm.controls.applicationLicenseId.enable();
        } else {
          this.deployForm.controls.applicationLicenseId.disable();
        }
      });

    this.deploymentTarget$
      .pipe(
        distinctUntilChanged((a, b) => a?.id === b?.id),
        takeUntil(this.destroyed$)
      )
      .subscribe((deploymentTarget) => {
        if (deploymentTarget) {
          if (deploymentTarget.type === 'kubernetes') {
            this.deployForm.controls.releaseName.enable();
            this.deployForm.controls.valuesYaml.enable();
            this.deployForm.controls.envFileData.disable();
            if (!this.deployForm.value.releaseName) {
              this.deployForm.patchValue({
                releaseName: deploymentTarget.name.trim().toLowerCase().replaceAll(/\W+/g, '-'),
              });
            }
          } else {
            this.deployForm.controls.envFileData.enable();
            this.deployForm.controls.releaseName.disable();
            this.deployForm.controls.valuesYaml.disable();
          }
          if (deploymentTarget.deployment) {
            this.deployForm.controls.applicationId.disable();
          } else {
            this.deployForm.controls.applicationId.enable();
          }
        }
      });

    combineLatest([
      this.applicationId$,
      this.applicationVersionId$,
      this.deploymentTarget$.pipe(
        distinctUntilChanged(
          (a, b) =>
            a?.id === b?.id &&
            a?.deployment?.envFileData === b?.deployment?.envFileData &&
            a?.deployment?.valuesYaml === b?.deployment?.valuesYaml
        )
      ),
    ])
      .pipe(
        debounceTime(5),
        switchMap(([applicationId, versionId, dt]) =>
          combineLatest([
            of(dt),
            // Only fill in the template if there is no existing deployment or the existing deployment has no values/env file
            versionId && applicationId && !(dt?.deployment?.valuesYaml || dt?.deployment?.envFileData)
              ? this.applications.getTemplateFile(applicationId, versionId).pipe(catchError(() => NEVER))
              : NEVER,
          ])
        ),
        takeUntil(this.destroyed$)
      )
      .subscribe(([deploymentTarget, templateFile]) => {
        if (deploymentTarget) {
          if (deploymentTarget.type === 'kubernetes') {
            this.deployForm.controls.valuesYaml.patchValue(templateFile ?? '');
          } else {
            this.deployForm.controls.envFileData.patchValue(templateFile ?? '');
          }
        }
      });

    this.licenses$.pipe(takeUntil(this.destroyed$)).subscribe((licenses) => {
      if (licenses.length > 0 && licenses[0].id) {
        this.deployForm.controls.applicationLicenseId.setValue(licenses[0].id);
      } else {
        this.deployForm.controls.applicationLicenseId.reset();
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
