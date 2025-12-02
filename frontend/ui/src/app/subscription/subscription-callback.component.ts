import {Component} from '@angular/core';
import {CommonModule} from '@angular/common';
import {FaIconComponent} from '@fortawesome/angular-fontawesome';
import {faCheckCircle} from '@fortawesome/free-solid-svg-icons';
import {RouterLink} from '@angular/router';

@Component({
  selector: 'app-subscription-callback',
  templateUrl: './subscription-callback.component.html',
  imports: [CommonModule, FaIconComponent, RouterLink],
})
export class SubscriptionCallbackComponent {
  protected readonly faCheckCircle = faCheckCircle;
}
