import {animate, state, style, transition, trigger} from '@angular/animations';

export const dropdownAnimation = trigger('dropdown', [
  state('in', style({transform: 'rotateX(0)'})),
  transition('void => *', [style({transform: 'rotateX(-90deg)'}), animate(100)]),
  transition('* => void', [animate(100, style({transform: 'rotateX(-90deg)'}))]),
]);
