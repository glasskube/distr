import {AnimationEvent} from '@angular/animations';
import {NgTemplateOutlet} from '@angular/common';
import {Component, HostBinding, HostListener, inject, OnInit, TemplateRef} from '@angular/core';
import {FormControl, ReactiveFormsModule} from '@angular/forms';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faXmark} from '@fortawesome/free-solid-svg-icons';
import {firstValueFrom, Subject} from 'rxjs';
import {modalFlyInOut} from '../../animations/modal';
import {AutotrimDirective} from '../../directives/autotrim.directive';
import {DialogRef, OverlayData} from '../../services/overlay.service';

export interface Message {
  message: string;
}

export interface Alert extends Message {
  type: 'warning' | 'danger';
}

export interface ConfirmMessage extends Message {
  alert?: Alert;
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

  protected readonly alertClass = [
    'p-4',
    'text-sm',
    'rounded-lg',
    ...(this.data.message?.alert?.type === 'warning'
      ? ['text-yellow-800', 'dark:text-yellow-300', 'bg-yellow-50', 'dark:bg-gray-800']
      : ['text-red-800', 'dark:text-red-400', 'bg-red-50', 'dark:bg-gray-800']),
  ];

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
