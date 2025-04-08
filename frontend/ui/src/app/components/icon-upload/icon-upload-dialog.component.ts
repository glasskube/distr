import {AnimationEvent} from '@angular/animations';
import {Component, HostBinding, HostListener, inject, OnInit, signal, TemplateRef} from '@angular/core';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faXmark} from '@fortawesome/free-solid-svg-icons';
import {firstValueFrom, map, Observable, Subject} from 'rxjs';
import {modalFlyInOut} from '../../animations/modal';
import {DialogRef, OverlayData} from '../../services/overlay.service';
import {AsyncPipe} from '@angular/common';
import {FormControl, FormGroup, ReactiveFormsModule} from '@angular/forms';
import {WithIcon} from '@glasskube/distr-sdk';
import {getFormDisplayedError} from '../../../util/errors';
import {ToastService} from '../../services/toast.service';


export interface IconUploadContext {
  data: WithIcon;
  customTemplate?: TemplateRef<any>;
  requiredConfirmInputText?: string;
}

@Component({
  imports: [FaIconComponent, ReactiveFormsModule, AsyncPipe],
  templateUrl: './icon-upload-dialog.component.html',
  animations: [modalFlyInOut],
})
export class IconUploadDialogComponent implements OnInit {
  public readonly faXmark = faXmark;
  public readonly dialogRef = inject(DialogRef) as DialogRef<boolean>;
  public readonly data = inject(OverlayData) as IconUploadContext;
  private readonly animationComplete$ = new Subject<void>();
  readonly confirmInput = new FormControl<string>('');

  private toast = inject(ToastService);

  protected readonly form = new FormGroup({
    icon: new FormControl<Blob | null>(null),
  });

  formLoading = signal(false);
  protected readonly iconSrc: Observable<string | null> = this.form.controls.icon.valueChanges.pipe(
    map((icon) => (icon ? URL.createObjectURL(icon) : null))
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


  onIconChange(event: Event) {
    const file = (event.target as HTMLInputElement).files?.[0];
    this.form.patchValue({icon: file ?? null});
  }

  deleteIcon() {
    this.form.patchValue({icon: null});
  }

  async save() {
    this.form.markAllAsTouched();
    if (this.form.valid) {
      this.formLoading.set(true);
      const formData = new FormData();
      formData.set('icon', this.form.value.icon ? (this.form.value.icon as File) : '');

      try {
        // this.organizationBranding = await lastValueFrom(req);
        this.toast.success('Branding saved successfully');
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
