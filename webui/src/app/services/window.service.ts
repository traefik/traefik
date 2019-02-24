import { Injectable } from '@angular/core';
import { EventManager } from '@angular/platform-browser';
import { Subject } from 'rxjs';
import { debounceTime } from 'rxjs/operators';

@Injectable()
export class WindowService {
  private readonly _resize = new Subject<EventTarget>();

  readonly resize = this._resize.asObservable();
  readonly resizeDebounce$ = this.resize.pipe(debounceTime(200));

  constructor(private eventManager: EventManager) {
    this.eventManager.addGlobalEventListener(
      'window',
      'resize',
      (event: UIEvent) => {
        this._resize.next(event.target);
      }
    );
  }
}
