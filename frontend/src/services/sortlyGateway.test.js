import { afterEach, describe, expect, it, vi } from 'vitest';
import { STATE_EVENT, createGateway, resolveBackend } from './sortlyGateway';

const METHODS = [
  ['getState', 'GetState', []],
  ['selectSource', 'SelectSource', []],
  ['selectDestination', 'SelectDestination', []],
  ['dropPaths', 'DropPaths', [['C:/a.txt']]],
  ['organize', 'Organize', []],
  ['undo', 'Undo', []],
  ['clearNotifications', 'ClearNotifications', []],
  ['setLanguage', 'SetLanguage', ['en']],
  ['setCriterion', 'SetCriterion', ['byDate', true]]
];

describe('resolveBackend', () => {
  afterEach(() => {
    delete window.go;
  });

  it('fora do Wails devolve null', () => {
    expect(resolveBackend({})).toBeNull();
  });

  it.each(METHODS)('%s chama o binding %s com os argumentos', async (method, binding, args) => {
    const fn = vi.fn().mockResolvedValue({ busy: '' });
    window.go = { app: { App: { [binding]: fn } } };
    await expect(resolveBackend(window)[method](...args)).resolves.toEqual({ busy: '' });
    expect(fn).toHaveBeenCalledWith(...args);
  });
});

describe('createGateway', () => {
  it('sem backend, as ações rejeitam e as assinaturas não fazem nada', async () => {
    const gateway = createGateway(() => null);
    await expect(gateway.organize()).rejects.toThrow('Backend indisponível');
    expect(gateway.subscribeState(vi.fn())).toBeTypeOf('function');
    expect(gateway.subscribeFileDrop(vi.fn())).toBeTypeOf('function');
  });

  it.each(METHODS)('%s repassa argumentos e resultado', async (method, _binding, args) => {
    const backend = { [method]: vi.fn().mockResolvedValue('estado') };
    await expect(createGateway(() => backend)[method](...args)).resolves.toBe('estado');
    expect(backend[method]).toHaveBeenCalledWith(...args);
  });
});

describe('assinaturas no Wails', () => {
  afterEach(() => {
    delete window.go;
    delete window.runtime;
  });

  it('estado: assina o evento sortly:state e devolve o cancelamento do Wails', () => {
    const off = vi.fn();
    window.go = { app: { App: {} } };
    window.runtime = { EventsOnMultiple: vi.fn(() => off) };
    const handler = vi.fn();

    const unsubscribe = createGateway().subscribeState(handler);

    expect(window.runtime.EventsOnMultiple).toHaveBeenCalledWith(STATE_EVENT, handler, -1);
    unsubscribe();
    expect(off).toHaveBeenCalled();
  });

  it('arquivos soltos: assina com alvo de drop e cancela com OnFileDropOff', () => {
    window.go = { app: { App: {} } };
    window.runtime = { OnFileDrop: vi.fn(), OnFileDropOff: vi.fn() };
    const handler = vi.fn();

    const unsubscribe = createGateway().subscribeFileDrop(handler);

    expect(window.runtime.OnFileDrop).toHaveBeenCalledWith(expect.any(Function), true);
    window.runtime.OnFileDrop.mock.calls[0][0](10, 20, ['C:/a.txt']);
    expect(handler).toHaveBeenCalledWith(['C:/a.txt']);
    unsubscribe();
    expect(window.runtime.OnFileDropOff).toHaveBeenCalled();
  });
});
