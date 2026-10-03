// Único ponto de acesso ao backend (ADR 0002). Na versão Wails, usa os
// bindings gerados em wailsjs/go/app/App.js; enquanto a versão Electron
// existir (até a remoção do Electron), usa window.electronAPI.
import * as WailsApp from '../../wailsjs/go/app/App';
import { OnFileDrop, OnFileDropOff } from '../../wailsjs/runtime/runtime';

export const UNEXPECTED = 'UNEXPECTED';

const wailsBackend = {
  selectSourceFolder: () => WailsApp.SelectSourceFolder(),
  selectDestinationFolder: () => WailsApp.SelectDestinationFolder(),
  resolveDroppedPath: (path) => WailsApp.ResolveDroppedPath(path),
  getLastOrganizationState: () => WailsApp.GetLastOrganizationState(),
  organizeFiles: (payload) => WailsApp.OrganizeFiles(payload),
  undoLastOrganization: () => WailsApp.UndoLastOrganization(),
  // O Wails entrega os caminhos só quando o drop termina num elemento com
  // --wails-drop-target: drop (useDropTarget = true).
  subscribeFileDrop: (handler) => {
    OnFileDrop((_x, _y, paths) => handler(paths), true);
    return () => OnFileDropOff();
  }
};

export function resolveBackend(win = window) {
  if (win.go?.app?.App) {
    return wailsBackend;
  }
  return win.electronAPI ?? null;
}

// Erro do backend normalizado: code é o código estável (ex.: NOTHING_TO_UNDO)
// quando existe; message guarda o texto original (versão Electron).
export class SortlyError extends Error {
  constructor(code, message) {
    super(message || code);
    this.name = 'SortlyError';
    this.code = code;
  }
}

const CODE_PATTERN = /^[A-Z][A-Z_]+$/;

// O Wails rejeita com o código como string ou como Error.message.
export function toSortlyError(error) {
  const message = typeof error === 'string' ? error : error?.message || '';
  const code = CODE_PATTERN.test(message) ? message : UNEXPECTED;
  return new SortlyError(code, message);
}

const noop = () => {};

export function createGateway(getBackend = resolveBackend) {
  const call = async (method, ...args) => {
    const backend = getBackend();
    if (!backend) {
      throw new SortlyError(UNEXPECTED, 'Backend indisponível');
    }
    try {
      return await backend[method](...args);
    } catch (error) {
      throw toSortlyError(error);
    }
  };

  return {
    selectSourceFolder: () => call('selectSourceFolder'),
    selectDestinationFolder: () => call('selectDestinationFolder'),
    resolveDroppedPath: (path) => call('resolveDroppedPath', path),
    getLastOrganizationState: () => call('getLastOrganizationState'),
    organizeFiles: (payload) => call('organizeFiles', payload),
    undoLastOrganization: () => call('undoLastOrganization'),
    // Assina os arquivos soltos na janela e devolve a função que cancela a
    // assinatura. Na versão Electron não há assinatura: o caminho vem do
    // próprio evento de drop (File.path).
    subscribeFileDrop: (handler) => getBackend()?.subscribeFileDrop?.(handler) ?? noop
  };
}

const sortlyGateway = createGateway();

export default sortlyGateway;
