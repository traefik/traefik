import { Pipe, PipeTransform } from '@angular/core';

@Pipe({
  name: 'frontendFilter',
  pure: false
})
export class FrontendFilterPipe implements PipeTransform {
  transform(items: any[], filter: string): any {
    if (!items || !filter) {
      return items;
    }

    const keyword = filter.toLowerCase();
    return items.filter(d => d.id.toLowerCase().includes(keyword)
      || d.backend.toLowerCase().includes(keyword)
      || d.routes.some(r => r.rule.toLowerCase().includes(keyword)));
  }
}
