import {AnimationEvent} from '@angular/animations';
import {Component, HostBinding, HostListener, inject, OnInit, TemplateRef} from '@angular/core';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faXmark} from '@fortawesome/free-solid-svg-icons';
import {firstValueFrom, Subject} from 'rxjs';
import {modalFlyInOut} from '../../animations/modal';
import {DialogRef, OverlayData} from '../../services/overlay.service';
import {NgTemplateOutlet} from '@angular/common';
import {AutotrimDirective} from '../../directives/autotrim.directive';
import {FormControl, ReactiveFormsModule} from '@angular/forms';

export interface Message {
  message: string;
}

export interface ConfirmMessage extends Message {
  warning?: Message;
}

export interface ConfirmConfig {
  message?: ConfirmMessage;
  customTemplate?: TemplateRef<any>;
  requiredConfirmInputText?: string;
}

@Component({
  imports: [FaIconComponent, NgTemplateOutlet, AutotrimDirective, ReactiveFormsModule],
  templateUrl: './confirm-dialog.component.html',
  animations: [modalFlyInOut],
})
export class ConfirmDialogComponent implements OnInit {
  public readonly faXmark = faXmark;
  public readonly dialogRef = inject(DialogRef) as DialogRef<boolean>;
  public readonly data = inject(OverlayData) as ConfirmConfig;
  private readonly animationComplete$ = new Subject<void>();
  readonly confirmInput = new FormControl<string>('');

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
}
