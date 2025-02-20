import {Directive, HostListener, inject} from '@angular/core';
import {NgControl} from '@angular/forms';

@Directive({selector: 'input[autotrim]'})
export class AutotrimDirective {
  private readonly ngControl = inject(NgControl, {optional: true});

  @HostListener('blur', ['$event']) onBlur(event: Event) {
    const target = event.target as HTMLInputElement;
    if (target.value !== target.value.trim()) {
      target.value = target.value.trim();
    }
    if (typeof this.ngControl?.value === 'string' && this.ngControl.value != this.ngControl.value.trim()) {
      this.ngControl.control?.setValue(this.ngControl.value.trim());
    }
  }
}
