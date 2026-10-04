import { describe, expect, it, vi } from 'vitest';
import { act, renderHook, waitFor } from '@testing-library/react';
import useViewState, { INITIAL_STATE } from './useViewState';

const state = (overrides = {}) => ({ ...INITIAL_STATE, ...overrides });

function fakeGateway(overrides = {}) {
  return {
    getState: vi.fn().mockResolvedValue(state({ sourceFolderPath: 'C:\\origem' })),
    selectSource: vi.fn().mockResolvedValue(state({ sourceFolderPath: 'C:\\nova' })),
    selectDestination: vi.fn().mockResolvedValue(state({ destinationFolderPath: 'C:\\destino' })),
    dropPaths: vi.fn().mockResolvedValue(state({ sourceFolderPath: 'C:\\solta' })),
    organize: vi.fn().mockResolvedValue(state({ hasUndo: true })),
    undo: vi.fn().mockResolvedValue(state({ hasUndo: false })),
    clearNotifications: vi.fn().mockResolvedValue(state()),
    setLanguage: vi.fn((language) =>
      Promise.resolve(state({ settings: { language, criteria: [] } }))
    ),
    setCriterion: vi.fn().mockResolvedValue(state()),
    subscribeState: vi.fn(() => vi.fn()),
    subscribeFileDrop: vi.fn(() => vi.fn()),
    ...overrides
  };
}

async function setup(gateway = fakeGateway()) {
  const hook = renderHook(() => useViewState(gateway));
  await waitFor(() => expect(hook.result.current.ready).toBe(true));
  return { ...hook, gateway };
}

describe('useViewState', () => {
  it('começa sem estado e fica pronto com o estado do backend', async () => {
    const gateway = fakeGateway();
    const { result } = renderHook(() => useViewState(gateway));
    expect(result.current.ready).toBe(false);
    expect(result.current.state).toBe(INITIAL_STATE);
    await waitFor(() => expect(result.current.ready).toBe(true));
    expect(result.current.state.sourceFolderPath).toBe('C:\\origem');
  });

  it('sem backend fica pronto com o estado inicial', async () => {
    const { result } = await setup(
      fakeGateway({ getState: vi.fn().mockRejectedValue(new Error('sem backend')) })
    );
    expect(result.current.state).toBe(INITIAL_STATE);
  });

  it.each([
    ['selectSource', [], { sourceFolderPath: 'C:\\nova' }],
    ['selectDestination', [], { destinationFolderPath: 'C:\\destino' }],
    ['organize', [], { hasUndo: true }],
    ['undo', [], { hasUndo: false }],
    ['setLanguage', ['en'], { settings: { language: 'en', criteria: [] } }]
  ])('%s mostra o estado devolvido', async (action, args, expected) => {
    const { result, gateway } = await setup();
    await act(() => result.current.actions[action](...args));
    expect(gateway[action]).toHaveBeenCalledWith(...args);
    expect(result.current.state).toMatchObject(expected);
  });

  it('limpar notificações e alterar critério chamam o backend', async () => {
    const { result, gateway } = await setup();
    await act(() => result.current.actions.clearNotifications());
    await act(() => result.current.actions.setCriterion('byDate', true));
    expect(gateway.clearNotifications).toHaveBeenCalled();
    expect(gateway.setCriterion).toHaveBeenCalledWith('byDate', true);
  });

  it('ação que falha mantém o estado anterior', async () => {
    const { result } = await setup(
      fakeGateway({ organize: vi.fn().mockRejectedValue(new Error('pânico')) })
    );
    await act(() => result.current.actions.organize());
    expect(result.current.state.sourceFolderPath).toBe('C:\\origem');
  });

  it('evento de estado substitui o estado e é cancelado ao desmontar', async () => {
    const unsubscribe = vi.fn();
    let onState;
    const gateway = fakeGateway({
      subscribeState: vi.fn((handler) => {
        onState = handler;
        return unsubscribe;
      })
    });
    const { result, unmount } = await setup(gateway);

    act(() => onState(state({ busy: 'organize' })));
    expect(result.current.state.busy).toBe('organize');

    unmount();
    expect(unsubscribe).toHaveBeenCalled();
  });

  it('arquivos soltos viram DropPaths', async () => {
    let onDrop;
    const gateway = fakeGateway({
      subscribeFileDrop: vi.fn((handler) => {
        onDrop = handler;
        return vi.fn();
      })
    });
    const { result } = await setup(gateway);

    await act(async () => onDrop(['C:\\solta\\a.txt']));

    expect(gateway.dropPaths).toHaveBeenCalledWith(['C:\\solta\\a.txt']);
    expect(result.current.state.sourceFolderPath).toBe('C:\\solta');
  });
});
