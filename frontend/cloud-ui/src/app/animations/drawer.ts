import {trigger, state, style, transition, animate} from '@angular/animations';

export const drawerFlyInOut = trigger('drawerFlyInOut', [
  state('in', style({transform: 'translateX(0)', opacity: '1'})),
  transition('void => *', [style({transform: 'translateX(100%)', opacity: '0'}), animate(150)]),
  transition('* => void', [animate(150, style({transform: 'translateX(100%)', opacity: '0'}))]),
]);
