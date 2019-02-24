import { Injectable } from '@angular/core';
import { EventManager } from '@angular/platform-browser';
import { Subject } from 'rxjs';

@Injectable()
export class WindowService {
  resize: Subject<any>;

  constructor(private eventManager: EventManager) {
    this.resize = new Subject();
    this.eventManager.addGlobalEventListener(
      'window',
      'resize',
      (event: UIEvent) => {
        this.resize.next(event.target);
      }
    );
  }
}
