import {Directive, ElementRef, HostListener, inject} from '@angular/core';

@Directive({selector: 'input[autotrim]'})
export class AutotrimDirective {
  @HostListener('blur', ['$event']) onBlur(event: Event) {
    const target = event.target as HTMLInputElement;
    if (target.value !== target.value.trim()) {
      target.value = target.value.trim();
    }
  }
}
