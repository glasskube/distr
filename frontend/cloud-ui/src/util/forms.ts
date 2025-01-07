import {FormGroup} from '@angular/forms';

export function enableControls(formGroup: FormGroup) {
  toggleControls(formGroup, true);
}

export function disableControls(formGroup: FormGroup) {
  toggleControls(formGroup, false);
}

export function toggleControls(formGroup: FormGroup, enabled: boolean) {
  for (let controlsKey in formGroup.controls) {
    if (enabled) {
      formGroup.controls[controlsKey].enable();
    } else {
      formGroup.controls[controlsKey].disable();
    }
  }
}
