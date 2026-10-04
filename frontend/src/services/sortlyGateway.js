// Único ponto de acesso ao backend (ADR 0002): usa os bindings do Wails
// gerados em wailsjs/go/app/App.js. Todo binding devolve o estado completo
// da tela (ADR 0005).
import * as WailsApp from '../../wailsjs/go/app/App';
import { EventsOn, OnFileDrop, OnFileDropOff } from '../../wailsjs/runtime/runtime';

export const STATE_EVENT = 'sortly:state';

const wailsBackend = {
  getState: () => WailsApp.GetState(),
  selectSource: () => WailsApp.SelectSource(),
  selectDestination: () => WailsApp.SelectDestination(),
  dropPaths: (paths) => WailsApp.DropPaths(paths),
  organize: () => WailsApp.Organize(),
  undo: () => WailsApp.Undo(),
  clearNotifications: () => WailsApp.ClearNotifications(),
  setLanguage: (language) => WailsApp.SetLanguage(language),
  setCriterion: (key, enabled) => WailsApp.SetCriterion(key, enabled),
  subscribeState: (handler) => EventsOn(STATE_EVENT, handler),
  // O Wails entrega os caminhos só quando o drop termina num elemento com
  // --wails-drop-target: drop (useDropTarget = true).
  subscribeFileDrop: (handler) => {
    OnFileDrop((_x, _y, paths) => handler(paths), true);
    return () => OnFileDropOff();
  }
};

// Fora da janela do Wails (ex.: vite dev no navegador), não há backend.
export function resolveBackend(win = window) {
  return win.go?.app?.App ? wailsBackend : null;
}

const noop = () => {};

export function createGateway(getBackend = resolveBackend) {
  const call = (method, ...args) => {
    const backend = getBackend();
    return backend ? backend[method](...args) : Promise.reject(new Error('Backend indisponível'));
  };
  const subscribe = (method, handler) => getBackend()?.[method]?.(handler) ?? noop;

  return {
    getState: () => call('getState'),
    selectSource: () => call('selectSource'),
    selectDestination: () => call('selectDestination'),
    dropPaths: (paths) => call('dropPaths', paths),
    organize: () => call('organize'),
    undo: () => call('undo'),
    clearNotifications: () => call('clearNotifications'),
    setLanguage: (language) => call('setLanguage', language),
    setCriterion: (key, enabled) => call('setCriterion', key, enabled),
    subscribeState: (handler) => subscribe('subscribeState', handler),
    subscribeFileDrop: (handler) => subscribe('subscribeFileDrop', handler)
  };
}

const sortlyGateway = createGateway();

export default sortlyGateway;
