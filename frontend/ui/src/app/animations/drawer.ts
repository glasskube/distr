import {animate, state, style, transition, trigger} from '@angular/animations';

export const drawerFlyInOut = trigger('drawerFlyInOut', [
  state('in', style({transform: 'translateX(0)', opacity: '1'})),
  transition('void => *', [style({transform: 'translateX(100%)', opacity: '0.8'}), animate('150ms ease-out')]),
  transition('* => void', [animate('150ms ease-in', style({transform: 'translateX(100%)', opacity: '0.8'}))]),
]);
