import {trigger, state, style, transition, animate} from '@angular/animations';

export const modalFlyInOut = trigger('modalFlyInOut', [
  state('in', style({transform: 'translateY(0)', opacity: '1'})),
  transition('void => *', [style({transform: 'translateY(-100%)', opacity: '0'}), animate(150)]),
  transition('* => void', [animate(150, style({transform: 'translateY(-100%)', opacity: '0'}))]),
]);
