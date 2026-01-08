import {Component, effect, input, output} from '@angular/core';
import {FormControl, FormGroup, ReactiveFormsModule} from '@angular/forms';
import {Application, ApplicationVersion} from '@glasskube/distr-sdk';
import {EditorComponent} from '../components/editor.component';

export interface ApplicationVersionDetail {
  application: Application;
  version: ApplicationVersion;
  linkTemplate: string;
  kubernetes?: {
    baseValues: string | null;
    template: string | null;
  };
  docker?: {
    compose: string | null;
    template: string | null;
  };
}

@Component({
  selector: 'app-application-version-detail-modal',
  templateUrl: './application-version-detail-modal.component.html',
  imports: [ReactiveFormsModule, EditorComponent],
})
export class ApplicationVersionDetailModalComponent {
  versionDetail = input.required<ApplicationVersionDetail>();
  closed = output<void>();

  versionDetailsForm = new FormGroup({
    name: new FormControl(''),
    link: new FormControl(''),
    kubernetes: new FormGroup({
      baseValues: new FormControl(''),
      template: new FormControl(''),
      linkTemplate: new FormControl(''),
    }),
    docker: new FormGroup({
      compose: new FormControl(''),
      template: new FormControl(''),
      linkTemplate: new FormControl(''),
    }),
  });

  constructor() {
    this.versionDetailsForm.disable();

    effect(() => {
      const detail = this.versionDetail();
      this.versionDetailsForm.patchValue({
        link: detail.linkTemplate,
        kubernetes: detail.kubernetes,
        docker: detail.docker,
      });
    });
  }

  close() {
    this.closed.emit();
  }
}
