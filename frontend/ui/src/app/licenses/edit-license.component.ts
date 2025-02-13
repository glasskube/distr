import {AfterViewInit, Component, forwardRef, inject, Injector, Input, OnDestroy, OnInit} from '@angular/core';
import {AsyncPipe} from '@angular/common';
import {AutotrimDirective} from '../directives/autotrim.directive';
import {
  ControlValueAccessor,
  FormArray,
  FormBuilder,
  FormControl,
  FormGroup,
  NG_VALUE_ACCESSOR,
  NgControl,
  ReactiveFormsModule,
  Validators,
  TouchedChangeEvent,
} from '@angular/forms';
import {faMagnifyingGlass, faPen, faPlus, faXmark} from '@fortawesome/free-solid-svg-icons';
import {firstValueFrom, map, Subject, switchMap, takeUntil} from 'rxjs';
import {ApplicationLicense} from '../types/application-license';
import {ApplicationsService} from '../services/applications.service';
import {Application, ApplicationVersion} from '../../../../../sdk/js/src';
import {UsersService} from '../services/users.service';

@Component({
  selector: 'app-edit-license',
  templateUrl: './edit-license.component.html',
  imports: [AsyncPipe, AutotrimDirective, ReactiveFormsModule],
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => EditLicenseComponent),
      multi: true,
    },
  ],
})
export class EditLicenseComponent implements OnInit, OnDestroy, AfterViewInit, ControlValueAccessor {
  private injector = inject(Injector);
  private readonly destroyed$ = new Subject<void>();
  private readonly applicationsService = inject(ApplicationsService);
  private readonly usersService = inject(UsersService);
  applications$ = this.applicationsService.list();
  customers$ = this.usersService.getUsers().pipe(map((accounts) => accounts.filter((a) => a.userRole === 'customer'))); // TODO cache users response
  private fb = inject(FormBuilder);
  license: ApplicationLicense | undefined;
  editForm = new FormGroup({
    id: new FormControl<string | undefined>(undefined, {nonNullable: true}),
    name: new FormControl<string | undefined>(undefined, {nonNullable: true, validators: Validators.required}),
    applicationId: new FormControl<string | undefined>(undefined, {nonNullable: true, validators: Validators.required}),
    includeAllVersions: new FormControl<boolean>(true, {
      nonNullable: true,
      validators: Validators.required,
    }),
    versions: this.fb.array<boolean>([]),
    ownerUserAccountId: new FormControl<string | undefined>(undefined, {nonNullable: true}),
    /*registryEnabled: new FormControl<boolean>(false),
    registry: new FormGroup({
      url: new FormControl('', Validators.required),
      username: new FormControl('', Validators.required),
      password: new FormControl('', Validators.required),
    }),*/
  });
  editFormLoading = false;
  selectedApplication: Application | undefined; // TODO fancy

  protected readonly faMagnifyingGlass = faMagnifyingGlass;

  ngOnInit() {
    // this.editForm.controls.registry.disable({emitEvent: false});
    this.editForm.valueChanges.pipe(takeUntil(this.destroyed$)).subscribe(() => {
      // this.onTouched();
      if (this.editForm.valid) {
        const val = this.editForm.getRawValue();
        this.onChange({
          id: val.id,
          name: val.name,
          applicationId: val.applicationId,
          versions: this.getSelectedVersions(val.includeAllVersions!, val.versions ?? []),
          ownerUserAccountId: val.ownerUserAccountId,
        });
      } else {
        this.onChange(undefined);
      }
    });
    this.editForm.controls.applicationId.valueChanges
      .pipe(
        takeUntil(this.destroyed$),
        switchMap(async (applicationId) => {
          const apps = await firstValueFrom(this.applicationsService.list());
          return apps.find((a) => a.id === applicationId);
        })
      )
      .subscribe((selectedApplication) => {
        this.selectedApplication = selectedApplication;
        this.versionsArray.clear({emitEvent: false});
        const versions = selectedApplication?.versions ?? [];
        for (let i = 0; i < versions.length; i++) {
          const version = versions[i];
          const selected = !!(this.license?.versions?.some((v) => v.id === version.id));
          /*
          business logic for now:
          * if license exists but not assigned yet:
              - owner is enabled
              - "all versions" and all app versions are enabled
              - application is disabled
              - registry stuff is enabled
          * if license assigned already:
              - owner is disabled
              - application is disabled
              - if "all versions" selected: "all versions" and all app versions disabled
              - if specific versions selected: "all versions" and unselected versions enabled; already selected versions disabled
              - registry stuff is enabled
           */
          this.versionsArray.push(this.fb.control(selected), {emitEvent: i === versions.length - 1});
        }
      });
  }

  private getSelectedVersions(includeAllVersions: boolean, versionControls: (boolean | null)[]): ApplicationVersion[] {
    if(includeAllVersions) {
      return [];
    }
    return versionControls.map((v, idx) => {
      if(v) {
        return this.selectedApplication?.versions?.[idx];
      }
      return undefined;
    }).filter(v => !!v);
  }

  ngAfterViewInit() {
    // from https://github.com/angular/angular/issues/45089
    this.injector
      .get(NgControl)
      .control!.events.pipe(takeUntil(this.destroyed$))
      .subscribe((event) => {
        if (event instanceof TouchedChangeEvent) {
          console.log('event', event);
          if(event.touched) {
            this.editForm.markAllAsTouched();
          }
        }
      });
  }

  ngOnDestroy() {
    this.destroyed$.next();
    this.destroyed$.complete();
  }

  get versionsArray() {
    return this.editForm.controls.versions as FormArray;
  }

  protected readonly faPlus = faPlus;
  protected readonly faXmark = faXmark;
  protected readonly faPen = faPen;

  private onChange: (l: ApplicationLicense | undefined) => void = () => {};
  private onTouched: () => void = () => {};

  registerOnChange(fn: (l: ApplicationLicense | undefined) => void): void {
    this.onChange = fn;
  }

  registerOnTouched(fn: any): void {
    this.onTouched = fn;
  }

  writeValue(license: ApplicationLicense | undefined): void {
    console.log('writeValue', license);
    this.license = license;
    if (license) {
      // TODO disable: applicationId if not assigned yet; assignee if assigned
      // TODO same on backend
      this.editForm.patchValue({
        id: license.id,
        name: license.name,
        applicationId: license.applicationId,
        versions: [], // will be set by applicationId-on-change,
        includeAllVersions: (license.versions ?? []).length === 0,
        ownerUserAccountId: license.ownerUserAccountId,
        /*registry: {
          url: license.registryUrl,
          username: license.registryUsername,
          password: license.registryPassword,
        },
        registryEnabled: !!(license.registryUrl && license.registryUsername && license.registryPassword),*/
      });
      if (license.ownerUserAccountId) {
        this.editForm.controls.applicationId.disable({emitEvent: false});
        this.editForm.controls.ownerUserAccountId.disable({emitEvent: false});
      }
      // TODO probably more disabling/enabling logic??
    } else {
      this.editForm.reset();
    }
  }
}
