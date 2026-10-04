import { afterEach, describe, expect, it, vi } from 'vitest';
import { SortlyError, createGateway, resolveBackend, toSortlyError } from './sortlyGateway';

describe('resolveBackend', () => {
  afterEach(() => {
    delete window.go;
  });

  it('usa os bindings do Wails quando presentes', async () => {
    window.go = {
      app: { App: { GetLastOrganizationState: vi.fn().mockResolvedValue({ hasUndo: true }) } }
    };
    const backend = resolveBackend(window);
    await expect(backend.getLastOrganizationState()).resolves.toEqual({ hasUndo: true });
  });

  it('fora do Wails devolve null', () => {
    expect(resolveBackend({})).toBeNull();
  });
});

describe('toSortlyError', () => {
  it.each([
    ['string com código (janela nativa)', 'NOTHING_TO_UNDO', 'NOTHING_TO_UNDO'],
    ['Error com código (navegador)', new Error('DROPPED_MISSING'), 'DROPPED_MISSING'],
    ['texto livre (erro do runtime)', new Error('TypeError: falha inesperada'), 'UNEXPECTED'],
    ['erro vazio', undefined, 'UNEXPECTED']
  ])('%s', (_, input, code) => {
    const error = toSortlyError(input);
    expect(error).toBeInstanceOf(SortlyError);
    expect(error.code).toBe(code);
  });
});

describe('createGateway', () => {
  it('sem backend rejeita com UNEXPECTED', async () => {
    const gateway = createGateway(() => null);
    await expect(gateway.selectSourceFolder()).rejects.toMatchObject({ code: 'UNEXPECTED' });
  });

  it('repassa argumentos e resultado', async () => {
    const backend = { organizeFiles: vi.fn().mockResolvedValue({ movedFiles: 2 }) };
    const gateway = createGateway(() => backend);
    await expect(gateway.organizeFiles('origem', 'destino')).resolves.toEqual({ movedFiles: 2 });
    expect(backend.organizeFiles).toHaveBeenCalledWith('origem', 'destino');
  });

  it('normaliza o erro do backend', async () => {
    const backend = { undoLastOrganization: vi.fn().mockRejectedValue('NOTHING_TO_UNDO') };
    const gateway = createGateway(() => backend);
    await expect(gateway.undoLastOrganization()).rejects.toMatchObject({ code: 'NOTHING_TO_UNDO' });
  });

  it('expõe os 9 métodos da API', () => {
    const backend = Object.fromEntries(
      [
        'selectSourceFolder',
        'selectDestinationFolder',
        'resolveDroppedPath',
        'getLastOrganizationState',
        'organizeFiles',
        'undoLastOrganization',
        'getSettings',
        'setLanguage',
        'setCriterion'
      ].map((m) => [m, vi.fn().mockResolvedValue(m)])
    );
    const gateway = createGateway(() => backend);
    return Promise.all(Object.keys(backend).map((m) => expect(gateway[m]('arg')).resolves.toBe(m)));
  });
});

describe('subscribeFileDrop', () => {
  afterEach(() => {
    delete window.go;
    delete window.runtime;
  });

  it('no Wails, assina com alvo de drop e cancela com OnFileDropOff', () => {
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

  it('sem backend, não assina nada', () => {
    expect(createGateway(() => null).subscribeFileDrop(vi.fn())).toBeTypeOf('function');
  });
});
