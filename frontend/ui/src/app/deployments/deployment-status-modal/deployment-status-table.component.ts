import {Component, input} from '@angular/core';
import {DatePipe} from '@angular/common';

export interface DeploymentStatusTableEntry {
  id?: string;
  date: string;
  status: string;
  detail: string;
}

@Component({
  selector: 'app-deployment-status-table',
  template: `<div class=" -mx-4 md:-mx-5 relative overflow-x-auto">
    <table class="w-full text-sm text-left rtl:text-right text-gray-500 dark:text-gray-400">
      <thead class="text-xs text-gray-700 uppercase bg-gray-50 dark:bg-gray-800 dark:text-gray-400 sr-only">
        <tr>
          <th scope="col">Date</th>
          <th scope="col">Status</th>
          <th scope="col">Details</th>
        </tr>
      </thead>
      <tbody>
        @for (entry of entries(); track entry.id && entry.date) {
          <tr class="not-last:border-b border-gray-200 dark:border-gray-600 hover:bg-gray-50 dark:hover:bg-gray-600">
            <th class="px-4 md:px-5 font-medium whitespace-nowrap">
              {{ entry.date | date: 'medium' }}
            </th>
            <td class="uppercase">
              {{ entry.status }}
            </td>
            <td class="px-4 md:px-5 whitespace-pre-line font-mono text-gray-900 dark:text-white">
              {{ entry.detail }}
            </td>
          </tr>
        }
      </tbody>
    </table>
  </div>`,
  imports: [DatePipe],
})
export class DeploymentStatusTableComponent {
  public readonly entries = input.required<DeploymentStatusTableEntry[]>();
}
