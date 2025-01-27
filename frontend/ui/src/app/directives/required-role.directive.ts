import {
  Directive,
  EmbeddedViewRef,
  inject,
  Input,
  OnChanges,
  SimpleChanges,
  TemplateRef,
  ViewContainerRef,
} from '@angular/core';
import {AuthService} from '../services/auth.service';
import {UserRole} from '@glasskube/distr-sdk';

@Directive({
  selector: '[appRequiredRole]',
})
export class RequireRoleDirective implements OnChanges {
  private readonly auth = inject(AuthService);
  private readonly templateRef = inject(TemplateRef);
  private readonly viewContainerRef = inject(ViewContainerRef);
  private embeddedViewRef: EmbeddedViewRef<unknown> | null = null;
  @Input({required: true, alias: 'appRequiredRole'}) public role!: UserRole;

  public ngOnChanges(changes: SimpleChanges): void {
    if (changes['role']) {
      if (this.auth.hasRole(this.role)) {
        if (this.embeddedViewRef === null) {
          this.embeddedViewRef = this.viewContainerRef.createEmbeddedView(this.templateRef);
        }
      } else {
        if (this.embeddedViewRef !== null) {
          this.embeddedViewRef.destroy();
          this.embeddedViewRef = null;
        }
      }
    }
  }
}
