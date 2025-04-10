import {AnimationEvent} from '@angular/animations';
import {Component, HostBinding, HostListener, inject, OnInit, signal, TemplateRef} from '@angular/core';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faXmark} from '@fortawesome/free-solid-svg-icons';
import {firstValueFrom, lastValueFrom, map, Observable, Subject} from 'rxjs';
import {modalFlyInOut} from '../../animations/modal';
import {DialogRef, OverlayData} from '../../services/overlay.service';
import {AsyncPipe} from '@angular/common';
import {FormControl, FormGroup, ReactiveFormsModule, Validators} from '@angular/forms';
import {Image, UserAccountWithRole} from '@glasskube/distr-sdk';
import {getFormDisplayedError} from '../../../util/errors';
import {ToastService} from '../../services/toast.service';
import {UsersService} from '../../services/users.service';


export interface ImageUploadContext {
  data: Image;
  type: 'user' | 'application';
}

@Component({
  imports: [FaIconComponent, ReactiveFormsModule, AsyncPipe],
  templateUrl: './image-upload-dialog.component.html',
  animations: [modalFlyInOut],
})
export class ImageUploadDialogComponent implements OnInit {
  public readonly faXmark = faXmark;
  public readonly dialogRef = inject(DialogRef) as DialogRef<boolean>;
  public readonly data = inject(OverlayData) as ImageUploadContext;
  private readonly animationComplete$ = new Subject<void>();

  private toast = inject(ToastService);
  private users = inject(UsersService)

  protected readonly form = new FormGroup({
    image: new FormControl<Blob | null>(null, Validators.required),
  });

  formLoading = signal(false);
  protected readonly imageSrc: Observable<string | null> = this.form.controls.image.valueChanges.pipe(
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

  async save() {
    this.form.markAllAsTouched();
    if (this.form.valid) {
      this.formLoading.set(true);
      const formData = new FormData();
      formData.set('image', this.form.value.image ? (this.form.value.image as File) : '');

      let uploadResult: Observable<void>;

      debugger;

      switch (this.data.type) {
        case 'user':
          uploadResult = this.users.patchImage(this.data.data.id!!, formData);
          break;
        // case 'application':
        //   break;
        default:
          this.toast.error('Unsupported image type');
          this.formLoading.set(false);
          return;
      }

      try {
        console.log('upload');
        await lastValueFrom(uploadResult);

        this.toast.success('Image saved successfully');
        await this.dialogRef.close(true);
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
