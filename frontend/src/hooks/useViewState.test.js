import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { act, renderHook, waitFor } from '@testing-library/react';
import * as Backend from '../../wailsjs/go/app/App';
import { EventsOn, OnFileDrop, OnFileDropOff } from '../../wailsjs/runtime/runtime';
import useViewState, { INITIAL_STATE, STATE_EVENT, newest } from './useViewState';

vi.mock('../../wailsjs/go/app/App', () => ({
  GetState: vi.fn(),
  SelectSource: vi.fn(),
  SelectDestination: vi.fn(),
  DropPaths: vi.fn(),
  Organize: vi.fn(),
  Undo: vi.fn(),
  ClearNotifications: vi.fn(),
  MarkNotificationsRead: vi.fn(),
  SetTheme: vi.fn(),
  SetLanguage: vi.fn(),
  SetCriterion: vi.fn()
}));

vi.mock('../../wailsjs/runtime/runtime', () => ({
  EventsOn: vi.fn(),
  OnFileDrop: vi.fn(),
  OnFileDropOff: vi.fn()
}));

// Cada estado criado é mais novo que o anterior, como no backend.
let lastVersion = 0;
const state = (overrides = {}) => ({ ...INITIAL_STATE, version: ++lastVersion, ...overrides });

const unsubscribeState = vi.fn();

beforeEach(() => {
  window.runtime = {};
  EventsOn.mockReturnValue(unsubscribeState);
  Backend.GetState.mockResolvedValue(state({ sourceFolderPath: 'C:\\origem' }));
});

afterEach(() => {
  delete window.runtime;
});

async function setup() {
  const hook = renderHook(() => useViewState());
  await waitFor(() => expect(hook.result.current.ready).toBe(true));
  return hook;
}

describe('newest', () => {
  const current = { version: 5 };
  it.each([
    ['versão maior substitui', { version: 6 }, true],
    ['mesma versão é ignorada', { version: 5 }, false],
    ['versão menor é ignorada', { version: 4 }, false],
    ['sem estado é ignorado', undefined, false]
  ])('%s', (_name, next, replaces) => {
    expect(newest(current, next)).toBe(replaces ? next : current);
  });
});

describe('useViewState', () => {
  it('começa sem estado e fica pronto com o estado do backend', async () => {
    const { result } = renderHook(() => useViewState());
    expect(result.current.ready).toBe(false);
    expect(result.current.state).toBe(INITIAL_STATE);
    await waitFor(() => expect(result.current.ready).toBe(true));
    expect(result.current.state.sourceFolderPath).toBe('C:\\origem');
  });

  it('binding que rejeita: fica pronto com o estado inicial', async () => {
    Backend.GetState.mockRejectedValue(new Error('pânico'));
    const { result } = await setup();
    expect(result.current.state).toBe(INITIAL_STATE);
  });

  // npm run dev no navegador: sem window.runtime, os bindings lançam erro
  // síncrono e as assinaturas do runtime não podem ser chamadas.
  it('fora do Wails: fica pronto com o estado inicial e não assina eventos', async () => {
    delete window.runtime;
    Backend.GetState.mockImplementation(() => {
      throw new TypeError("Cannot read properties of undefined (reading 'app')");
    });
    const { result } = await setup();
    expect(result.current.state).toBe(INITIAL_STATE);
    expect(EventsOn).not.toHaveBeenCalled();
    expect(OnFileDrop).not.toHaveBeenCalled();
  });

  it.each([
    ['selectSource', 'SelectSource', [], { sourceFolderPath: 'C:\\nova' }],
    ['selectDestination', 'SelectDestination', [], { destinationFolderPath: 'C:\\destino' }],
    ['organize', 'Organize', [], { hasUndo: true }],
    ['undo', 'Undo', [], { hasUndo: false }],
    ['clearNotifications', 'ClearNotifications', [], { notifications: [] }],
    ['markNotificationsRead', 'MarkNotificationsRead', [], { unread: false }],
    ['setLanguage', 'SetLanguage', ['en'], { settings: { language: 'en', criteria: [] } }],
    ['setTheme', 'SetTheme', ['light'], { settings: { theme: 'light', criteria: [] } }],
    ['setCriterion', 'SetCriterion', ['byDate', true], { busy: '' }]
  ])('%s chama %s e mostra o estado devolvido', async (action, binding, args, expected) => {
    const { result } = await setup();
    Backend[binding].mockResolvedValue(state(expected));
    await act(() => result.current.actions[action](...args));
    expect(Backend[binding]).toHaveBeenCalledWith(...args);
    expect(result.current.state).toMatchObject(expected);
  });

  it('ação que falha mantém o estado anterior', async () => {
    const { result } = await setup();
    Backend.Organize.mockRejectedValue(new Error('pânico'));
    await act(() => result.current.actions.organize());
    expect(result.current.state.sourceFolderPath).toBe('C:\\origem');
  });

  it('evento de estado substitui o estado e é cancelado ao desmontar', async () => {
    const { result, unmount } = await setup();
    expect(EventsOn).toHaveBeenCalledWith(STATE_EVENT, expect.any(Function));
    const onState = EventsOn.mock.calls[0][1];

    act(() => onState(state({ busy: 'organize' })));
    expect(result.current.state.busy).toBe('organize');

    unmount();
    expect(unsubscribeState).toHaveBeenCalled();
  });

  // Regressão (#55): o "organizando" atrasado não pode travar a tela.
  it('estado antigo que chega depois não sobrescreve o mais novo', async () => {
    const busy = state({ busy: 'organize' });
    const done = state({ busy: '', hasUndo: true });
    const { result } = await setup();
    Backend.Organize.mockResolvedValue(done);
    const onState = EventsOn.mock.calls[0][1];

    await act(() => result.current.actions.organize());
    act(() => onState(busy));

    expect(result.current.state).toBe(done);
  });

  it('arquivos soltos no alvo viram DropPaths; a assinatura é cancelada ao desmontar', async () => {
    Backend.DropPaths.mockResolvedValue(state({ sourceFolderPath: 'C:\\solta' }));
    const { result, unmount } = await setup();
    expect(OnFileDrop).toHaveBeenCalledWith(expect.any(Function), true);
    const onDrop = OnFileDrop.mock.calls[0][0];

    await act(async () => onDrop(10, 20, ['C:\\solta\\a.txt']));

    expect(Backend.DropPaths).toHaveBeenCalledWith(['C:\\solta\\a.txt']);
    expect(result.current.state.sourceFolderPath).toBe('C:\\solta');
    unmount();
    expect(OnFileDropOff).toHaveBeenCalled();
  });
});
