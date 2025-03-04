import {
  AfterViewInit,
  Component,
  effect,
  ElementRef,
  forwardRef,
  inject,
  Injector,
  OnDestroy,
  OnInit,
  signal,
  ViewChild,
} from '@angular/core';
import {AsyncPipe} from '@angular/common';
import {AutotrimDirective} from '../../directives/autotrim.directive';
import {
  AbstractControl,
  ControlValueAccessor,
  FormArray,
  FormBuilder,
  FormGroup,
  NG_VALUE_ACCESSOR,
  NgControl,
  ReactiveFormsModule,
  TouchedChangeEvent,
  Validators,
} from '@angular/forms';
import {faChevronDown, faMagnifyingGlass, faPen, faPlus, faXmark} from '@fortawesome/free-solid-svg-icons';
import {first, firstValueFrom, map, Subject, switchMap, takeUntil} from 'rxjs';
import {ApplicationLicense} from '../../types/application-license';
import {ApplicationsService} from '../../services/applications.service';
import {UsersService} from '../../services/users.service';
import dayjs from 'dayjs';
import {CdkConnectedOverlay, CdkOverlayOrigin} from '@angular/cdk/overlay';
import {dropdownAnimation} from '../../animations/dropdown';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {ArtifactLicense, ArtifactLicenseSelection} from '../../services/artifact-licenses.service';
import {ArtifactsService, ArtifactTag, ArtifactWithTags} from '../../services/artifacts.service';
import {ArtifactsHashComponent} from '../components';
import {RelativeDatePipe} from '../../../util/dates';

@Component({
  selector: 'app-edit-artifact-license',
  templateUrl: './edit-artifact-license.component.html',
  imports: [
    AsyncPipe,
    AutotrimDirective,
    ReactiveFormsModule,
    CdkOverlayOrigin,
    CdkConnectedOverlay,
    FaIconComponent,
    ArtifactsHashComponent,
    RelativeDatePipe,
  ],
  providers: [
    {
      provide: NG_VALUE_ACCESSOR,
      useExisting: forwardRef(() => EditArtifactLicenseComponent),
      multi: true,
    },
  ],
  animations: [dropdownAnimation],
})
export class EditArtifactLicenseComponent implements OnInit, OnDestroy, AfterViewInit, ControlValueAccessor {
  private injector = inject(Injector);
  private readonly destroyed$ = new Subject<void>();
  private readonly artifactsService = inject(ArtifactsService);
  private readonly usersService = inject(UsersService);
  allArtifacts$ = this.artifactsService.list();
  customers$ = this.usersService.getUsers().pipe(
    map((accounts) => accounts.filter((a) => a.userRole === 'customer')),
    first()
  );

  private fb = inject(FormBuilder);
  editForm = this.fb.nonNullable.group({
    id: this.fb.nonNullable.control<string | undefined>(undefined),
    name: this.fb.nonNullable.control<string | undefined>(undefined, Validators.required),
    expiresAt: this.fb.nonNullable.control(''),
    artifacts: this.fb.array([]),
    /*subjectId: this.fb.nonNullable.control<string | undefined>(undefined, Validators.required),
    includeAllItems: this.fb.nonNullable.control<boolean>(true, Validators.required),
    subjectItems: this.fb.array<boolean>([]),*/
    ownerUserAccountId: this.fb.nonNullable.control<string | undefined>(undefined),
  });
  editFormLoading = false;
  readonly license = signal<ArtifactLicense | undefined>(undefined);
  readonly selectedSubject = signal<ArtifactWithTags | undefined>(undefined);

  readonly openedArtifactIdx = signal<number | undefined>(undefined);
  dropdownOpen = signal(false);
  protected subjectItemsSelected = 0;

  dropdownWidth: number = 0;

  protected readonly faMagnifyingGlass = faMagnifyingGlass;
  protected readonly faChevronDown = faChevronDown;
  protected readonly faPlus = faPlus;
  protected readonly faXmark = faXmark;
  protected readonly faPen = faPen;

  @ViewChild('dropdownTriggerButton') dropdownTriggerButton!: ElementRef<HTMLElement>;

  constructor() {
    /*effect(() => {
      if (!this.openedArtifactIdx()) {
        if (
          !this.editForm.controls.includeAllItems.value &&
          !this.editForm.controls.subjectItems.value.some((v) => !!v)
        ) {
          this.editForm.controls.includeAllItems.patchValue(true);
        }
      }
    });*/
  }

  ngOnInit() {
    this.editForm.valueChanges.pipe(takeUntil(this.destroyed$)).subscribe(() => {
      this.onTouched();
      const val = this.editForm.getRawValue();
      console.log(val);
      /*if (!val.includeAllItems) {
        this.subjectItemsSelected = val.subjectItems.filter((v) => !!v).length;
      }
      if (this.editForm.valid) {
        this.onChange({
          id: val.id,
          name: val.name,
          expiresAt: val.expiresAt ? new Date(val.expiresAt) : undefined,
          artifactId: val.subjectId,
          artifactTags: this.getSelectedTags(val.includeAllItems!, val.subjectItems ?? []),
          ownerUserAccountId: val.ownerUserAccountId,
        });
      } else {
        this.onChange(undefined);
      }*/
    });

  }

  selectedArtifact(): ArtifactWithTags | undefined {
    return this.selectedSubject() as ArtifactWithTags;
  }

  private getSelectedTags(includeAllVersions: boolean, itemControls: (boolean | null)[]): ArtifactTag[] {
    if (includeAllVersions) {
      return [];
    }
    const artifact = this.selectedArtifact();
    return itemControls
      .map((v, idx) => {
        if (v) {
          return artifact?.tags?.[idx];
        }
        return undefined;
      })
      .filter((v) => !!v);
  }

  ngAfterViewInit() {
    // from https://github.com/angular/angular/issues/45089
    this.injector
      .get(NgControl)
      .control!.events.pipe(takeUntil(this.destroyed$))
      .subscribe((event) => {
        if (event instanceof TouchedChangeEvent) {
          if (event.touched) {
            this.editForm.markAllAsTouched();
          }
        }
      });
  }

  toggleDropdown(i: number) {
    this.openedArtifactIdx.update((idx) => idx === i ? undefined : i)
    // this.dropdownOpen.update((v) => !v);
    if (this.openedArtifactIdx()) {
      this.dropdownWidth = this.dropdownTriggerButton.nativeElement.getBoundingClientRect().width;
    }
  }

  ngOnDestroy() {
    this.destroyed$.next();
    this.destroyed$.complete();
  }

  get artifacts() {
    return this.editForm.controls.artifacts as FormArray;
  }

  asFormArray(control: AbstractControl): FormArray {
    return control as FormArray;
  }

  addArtifactGroup(selection?: ArtifactLicenseSelection) {
    const artifactGroup = this.fb.group({
      artifactId: this.fb.nonNullable.control<string | undefined>(selection?.artifact.id, Validators.required),
      includeAllTags: this.fb.nonNullable.control<boolean>((selection?.tags || []).length === 0, Validators.required),
      artifactTags: this.fb.array<boolean>([]), // TODO
    });
    artifactGroup.controls.includeAllTags.valueChanges.pipe(takeUntil(this.destroyed$)).subscribe((includeAll) => {
      if (includeAll) {
        artifactGroup.controls.artifactTags.controls.forEach((c) => c.patchValue(false, {emitEvent: false}));
      }
    });
    artifactGroup.controls.artifactTags.valueChanges.pipe(takeUntil(this.destroyed$)).subscribe((val) => {
      if (artifactGroup.controls.includeAllTags.value && val.some((v) => !!v)) {
        artifactGroup.controls.includeAllTags.patchValue(false, {emitEvent: false});
      }
    });
    artifactGroup.controls.artifactId.valueChanges
      .pipe(
        takeUntil(this.destroyed$),
        switchMap(async (artifactId) => {
          const artifacts = await firstValueFrom(this.artifactsService.list());
          return artifacts.find((a) => a.id === artifactId);
        })
      )
      .subscribe((selectedArtifact) => {
        artifactGroup.controls.artifactTags.clear({emitEvent: false});
        const allTagsOfArtifact = (selectedArtifact as ArtifactWithTags)?.tags ?? [];
        const licenseItems = this.license()?.artifacts?.find(a => a.artifact.id === selectedArtifact?.id)?.tags;
        let anySelected = false;
        for (let i = 0; i < allTagsOfArtifact.length; i++) {
          const item = allTagsOfArtifact[i];
          const selected = !!licenseItems?.some((v) => v.id === item.id);
          artifactGroup.controls.artifactTags.push(this.fb.control(selected), {emitEvent: i === allTagsOfArtifact.length - 1});
          anySelected = anySelected || selected;
        }
        if (!anySelected) {
          artifactGroup.controls.includeAllTags.patchValue(true);
        }
        this.selectedSubject.set(selectedArtifact);
      });
    this.artifacts.push(artifactGroup);
  }

  deleteArtifactGroup(i: number) {
    console.log('delete ', i)
    // TODO seems to be working incorrectly
    this.artifacts.removeAt(i);
  }

  private onChange: (l: ApplicationLicense | ArtifactLicense | undefined) => void = () => {};
  private onTouched: () => void = () => {};

  registerOnChange(fn: (l: ApplicationLicense | ArtifactLicense | undefined) => void): void {
    this.onChange = fn;
  }

  registerOnTouched(fn: any): void {
    this.onTouched = fn;
  }

  writeValue(license: ArtifactLicense | undefined): void {
    this.license.set(license);
    if (license) {
      this.editForm.patchValue({
        id: license.id,
        name: license.name,
        artifacts: [],
        expiresAt: license.expiresAt ? dayjs(license.expiresAt).format('YYYY-MM-DD') : '',
        ownerUserAccountId: license.ownerUserAccountId,
      });
      for(let selection of license.artifacts || []) {
        this.addArtifactGroup(selection);
      }
      if (license.ownerUserAccountId) {
        this.editForm.controls.artifacts.disable({emitEvent: false});
        this.editForm.controls.ownerUserAccountId.disable({emitEvent: false});
      }
    } else {
      this.editForm.reset();
    }
  }
}
