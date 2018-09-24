import { Pipe, PipeTransform } from '@angular/core';

/**
 * HumanReadableFilterPipe converts a time period in nanoseconds to a human-readable
 * string.
 */
@Pipe({name: 'humanreadable'})
export class HumanReadableFilterPipe implements PipeTransform {
  transform(value): any {
    let result = '';
    const powerOf10 = Math.floor(Math.log10(value));

    if (powerOf10 > 11) {
      result = value / (60 * Math.pow(10, 9)) + 'm';
    } else if (powerOf10 > 9) {
      result = value / Math.pow(10, 9) + 's';
    } else if (powerOf10 > 6) {
      result = value / Math.pow(10, 6) + 'ms';
    } else if (value > 0) {
      result = Math.floor(value) + 'ns';
    } else {
      result = value;
    }

    return result;
  }
}
