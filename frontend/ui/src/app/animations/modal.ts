import {trigger, state, style, transition, animate} from '@angular/animations';

export const modalFlyInOut = trigger('modalFlyInOut', [
  state('visible', style({transform: 'translateY(0)', opacity: '1'})),
  state('hidden', style({transform: 'translateY(-100%)', opacity: '0.8'})),
  transition('void => *', [style({transform: 'translateY(-100%)', opacity: '0.8'}), animate('150ms ease-out')]),
  transition('* => void', [animate('150ms ease-in', style({transform: 'translateY(-100%)', opacity: '0.8'}))]),
  transition('visible <=> hidden', [animate('150ms ease-in')]),
]);
