import {Component, inject} from '@angular/core';
import {modalFlyInOut} from '../../animations/modal';
import {DialogRef, OverlayData} from '../../services/overlay.service';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faXmark} from '@fortawesome/free-solid-svg-icons';

@Component({
  imports: [FaIconComponent],
  templateUrl: './confirm-dialog.component.html',
  animations: [modalFlyInOut],
})
export class ConfirmDialogComponent {
  public readonly faXmark = faXmark;
  public readonly dialogRef = inject(DialogRef) as DialogRef<boolean>;
  public readonly data = inject(OverlayData) as {message: string};
}
