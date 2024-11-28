import {GlobalPositionStrategy, Overlay, OverlayRef} from '@angular/cdk/overlay';
import {TemplatePortal} from '@angular/cdk/portal';
import {EmbeddedViewRef, inject, Injectable, TemplateRef, ViewContainerRef} from '@angular/core';
import {Observable, Subject, takeUntil} from 'rxjs';

export class ModalRef {
  private readonly closed$ = new Subject<void>();

  constructor(private readonly embeddedViewRef: EmbeddedViewRef<unknown>) {}

  public close() {
    this.embeddedViewRef.destroy();
    this.closed$.next();
  }

  public closed(): Observable<void> {
    return this.closed$;
  }
}

@Injectable({providedIn: 'root'})
export class ModalService {
  private readonly overlay = inject(Overlay);

  /**
   * @param templateRef the template to show
   * @param viewContainerRef needed to create a TemplatePortal. you can get it by injecting `ViewContainerRef`
   * @returns a handle of the modal with some control functions
   */
  public show(templateRef: TemplateRef<unknown>, viewContainerRef: ViewContainerRef): ModalRef {
    const overlayRef = this.overlay.create({
      hasBackdrop: true,
      positionStrategy: new GlobalPositionStrategy().centerHorizontally().top(),
    });
    const modalRef = new ModalRef(overlayRef.attach(new TemplatePortal(templateRef, viewContainerRef)));
    overlayRef
      .backdropClick()
      .pipe(takeUntil(modalRef.closed()))
      .subscribe(() => modalRef.close());
    return modalRef;
  }
}
