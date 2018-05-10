import { Pipe, PipeTransform } from '@angular/core';

@Pipe({
  name: 'backendFilter',
  pure: false
})
export class BackendFilterPipe implements PipeTransform {
  transform(items: any[], filter: string): any {
    if (!items || !filter) {
      return items;
    }

    const keyword = filter.toLowerCase();
    return items.filter(d => d.id.toLowerCase().includes(keyword)
      || d.servers.some(r => r.url.toLowerCase().includes(keyword)));
  }
}
