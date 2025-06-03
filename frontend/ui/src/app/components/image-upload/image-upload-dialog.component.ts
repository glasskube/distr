import {AnimationEvent} from '@angular/animations';
import {Component, HostBinding, HostListener, inject, OnDestroy, OnInit, signal} from '@angular/core';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faXmark} from '@fortawesome/free-solid-svg-icons';
import {firstValueFrom, lastValueFrom, map, Observable, Subject, takeUntil} from 'rxjs';
import {modalFlyInOut} from '../../animations/modal';
import {DialogRef, OverlayData} from '../../services/overlay.service';
import {AsyncPipe} from '@angular/common';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {getFormDisplayedError} from '../../../util/errors';
import {ToastService} from '../../services/toast.service';
import {FileScope, FilesService} from '../../services/files.service';

export interface ImageUploadContext {
  scope?: FileScope;
  imageUrl?: string;
}

@Component({
  imports: [FaIconComponent, ReactiveFormsModule, AsyncPipe],
  templateUrl: './image-upload-dialog.component.html',
  animations: [modalFlyInOut],
})
export class ImageUploadDialogComponent implements OnInit, OnDestroy {
  public readonly faXmark = faXmark;
  public readonly dialogRef = inject(DialogRef) as DialogRef<string>;
  public readonly data = inject(OverlayData) as ImageUploadContext;
  private readonly animationComplete$ = new Subject<void>();
  private readonly destroyed$ = new Subject<void>();
  private toast = inject(ToastService);

  private files = inject(FilesService);
  protected readonly form = new FormGroup({
    image: new FormControl<Blob | null>(null, Validators.required),
  });

  formLoading = signal(false);

  protected readonly imageSrc: Observable<string | null> = this.form.controls.image.valueChanges.pipe(
    takeUntil(this.destroyed$),
    map((image) => (image ? URL.createObjectURL(image) : null))
  );
  @HostBinding('@modalFlyInOut') public animationState = 'visible';

  @HostListener('@modalFlyInOut.done', ['$event']) onAnimationComplete(event: AnimationEvent) {
    if (event.toState === 'hidden') {
      this.animationComplete$.next();
    }
  }

  ngOnInit(): void {
    this.dialogRef.addOnClosedHook(async () => {
      this.animationState = 'hidden';
      await firstValueFrom(this.animationComplete$);
    });
  }

  onImageChange(event: Event) {
    const file = (event.target as HTMLInputElement).files?.[0];
    this.form.patchValue({image: file ?? null});
  }

  deleteImage() {
    this.form.patchValue({image: null});
  }

  ngOnDestroy(): void {
    this.destroyed$.next();
    this.destroyed$.complete();
  }

  async save() {
    this.form.markAllAsTouched();
    if (this.form.valid) {
      this.formLoading.set(true);
      const formData = new FormData();
      formData.set('file', this.form.value.image as File);

      try {
        let uploadResult = this.files.uploadFile(formData, this.data.scope);
        await this.dialogRef.close(await lastValueFrom(uploadResult));
        this.toast.success('Image saved successfully');
      } catch (e) {
        const msg = getFormDisplayedError(e);
        if (msg) {
          this.toast.error(msg);
        }
      } finally {
        this.formLoading.set(false);
      }
    }
  }
}
