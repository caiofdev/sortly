import { describe, expect, it, vi } from 'vitest';
import { act, renderHook, waitFor } from '@testing-library/react';
import { DEFAULT_OPTIONS } from '../domain/organizationOptions';
import { SortlyError } from '../services/sortlyGateway';
import useFileOrganizerController from './useFileOrganizerController';

function fakeGateway(overrides = {}) {
  return {
    getLastOrganizationState: vi.fn().mockResolvedValue({ hasUndo: false }),
    subscribeFileDrop: vi.fn(() => vi.fn()),
    selectSourceFolder: vi.fn().mockResolvedValue('C:\\origem'),
    selectDestinationFolder: vi.fn().mockResolvedValue(''),
    resolveDroppedPath: vi.fn().mockResolvedValue({ sourceFolderPath: 'C:\\solta' }),
    organizeFiles: vi.fn().mockResolvedValue({
      canUndo: true,
      movedFiles: 1,
      sourceFolderPath: 'C:\\origem',
      destinationFolderPath: 'C:\\origem',
      processedFiles: 1,
      ignoredWithoutExtension: 0,
      ignoredFolders: 0,
      failedFiles: 0,
      unchangedFiles: 0
    }),
    undoLastOrganization: vi.fn().mockResolvedValue({
      canUndo: false,
      restoredFiles: 1,
      renamedOnRestore: 0,
      skippedMissing: 0,
      failedFiles: 0
    }),
    ...overrides
  };
}

async function setup(gateway = fakeGateway(), language = 'pt-BR') {
  const notify = vi.fn();
  const hook = renderHook(() =>
    useFileOrganizerController({ language, organizationOptions: DEFAULT_OPTIONS, notify, gateway })
  );
  await waitFor(() => expect(gateway.getLastOrganizationState).toHaveBeenCalled());
  return { ...hook, notify, gateway };
}

describe('estado inicial', () => {
  it('recupera a última organização e avisa', async () => {
    const gateway = fakeGateway({
      getLastOrganizationState: vi.fn().mockResolvedValue({
        hasUndo: true,
        sourceFolderPath: 'C:\\a',
        destinationFolderPath: 'C:\\b'
      })
    });
    const { result, notify } = await setup(gateway);
    await waitFor(() => expect(result.current.hasUndo).toBe(true));
    expect(result.current.sourceFolderPath).toBe('C:\\a');
    expect(result.current.destinationFolderPath).toBe('C:\\b');
    expect(notify).toHaveBeenCalledWith(
      'info',
      expect.stringContaining('Última organização recuperada')
    );
  });

  it('erro do backend deixa sem desfazer e sem aviso', async () => {
    const gateway = fakeGateway({
      getLastOrganizationState: vi.fn().mockRejectedValue(new SortlyError('UNEXPECTED'))
    });
    const { result, notify } = await setup(gateway);
    expect(result.current.hasUndo).toBe(false);
    expect(notify).not.toHaveBeenCalled();
  });
});

describe('seleção de pastas', () => {
  it('escolhida, cancelada e com erro', async () => {
    const gateway = fakeGateway({
      selectDestinationFolder: vi.fn().mockRejectedValue(new Error('x'))
    });
    const { result, notify } = await setup(gateway);

    await act(() => result.current.handleSelectSourceFolder());
    expect(result.current.sourceFolderPath).toBe('C:\\origem');

    gateway.selectSourceFolder.mockResolvedValue('');
    await act(() => result.current.handleSelectSourceFolder());
    expect(result.current.sourceFolderPath).toBe('C:\\origem'); // cancelar não apaga

    await act(() => result.current.handleSelectDestinationFolder());
    expect(notify).toHaveBeenCalledWith('error', 'Não foi possível selecionar a pasta de destino.');
  });
});

async function drop(gateway, paths) {
  const onDrop = gateway.subscribeFileDrop.mock.calls.at(-1)[0];
  await act(async () => onDrop(paths));
}

describe('arrastar e soltar', () => {
  it('define a origem e avisa', async () => {
    const { result, notify, gateway } = await setup();
    await drop(gateway, ['C:\\solta\\a.txt']);
    await waitFor(() => expect(result.current.sourceFolderPath).toBe('C:\\solta'));
    expect(gateway.resolveDroppedPath).toHaveBeenCalledWith('C:\\solta\\a.txt');
    expect(notify).toHaveBeenCalledWith('info', 'Origem definida por arrastar e soltar. C:\\solta');
  });

  it('sem pasta no resultado não muda nada', async () => {
    const { result, notify, gateway } = await setup(
      fakeGateway({ resolveDroppedPath: vi.fn().mockResolvedValue({}) })
    );
    await drop(gateway, ['x']);
    await waitFor(() => expect(gateway.resolveDroppedPath).toHaveBeenCalled());
    expect(result.current.sourceFolderPath).toBe('');
    expect(notify).not.toHaveBeenCalled();
  });

  it('erro com código é traduzido', async () => {
    const gateway = fakeGateway({
      resolveDroppedPath: vi.fn().mockRejectedValue(new SortlyError('DROPPED_MISSING'))
    });
    const { notify } = await setup(gateway, 'en');
    await drop(gateway, ['x']);
    await waitFor(() =>
      expect(notify).toHaveBeenCalledWith('error', 'Dropped item no longer exists.')
    );
  });

  it('usa o primeiro item, ignora lista vazia e cancela ao desmontar', async () => {
    const unsubscribe = vi.fn();
    const gateway = fakeGateway({ subscribeFileDrop: vi.fn(() => unsubscribe) });
    const { result, unmount } = await setup(gateway);

    await drop(gateway, ['C:/fotos/a.jpg', 'C:/docs/b.pdf']);
    expect(gateway.resolveDroppedPath).toHaveBeenCalledWith('C:/fotos/a.jpg');
    await waitFor(() => expect(result.current.sourceFolderPath).toBe('C:\\solta'));

    await drop(gateway, []);
    await drop(gateway, null);
    expect(gateway.resolveDroppedPath).toHaveBeenCalledTimes(1);

    unmount();
    expect(unsubscribe).toHaveBeenCalled();
  });
});

describe('organizar e desfazer', () => {
  it('sem origem avisa e não chama o backend', async () => {
    const { result, notify, gateway } = await setup();
    await act(() => result.current.handleOrganizeFiles());
    expect(notify).toHaveBeenCalledWith('error', 'Selecione a pasta de origem antes de organizar.');
    expect(gateway.organizeFiles).not.toHaveBeenCalled();
  });

  it('organiza com destino vazio usando a origem e habilita desfazer', async () => {
    const { result, notify, gateway } = await setup();
    await act(() => result.current.handleSelectSourceFolder());
    await act(() => result.current.handleOrganizeFiles());
    expect(gateway.organizeFiles).toHaveBeenCalledWith({
      sourceFolderPath: 'C:\\origem',
      destinationFolderPath: 'C:\\origem',
      organizationOptions: DEFAULT_OPTIONS
    });
    expect(result.current.hasUndo).toBe(true);
    expect(result.current.loadingAction).toBeNull();
    expect(notify).toHaveBeenCalledWith(
      'organize',
      expect.stringContaining('1 arquivo(s) movido(s)')
    );
  });

  it('erro ao organizar é traduzido', async () => {
    const gateway = fakeGateway({
      organizeFiles: vi.fn().mockRejectedValue(new SortlyError('NO_CRITERIA'))
    });
    const { result, notify } = await setup(gateway);
    await act(() => result.current.handleSelectSourceFolder());
    await act(() => result.current.handleOrganizeFiles());
    expect(notify).toHaveBeenCalledWith('error', 'Selecione ao menos um criterio de organizacao.');
  });

  it('desfazer com sucesso e sem nada para desfazer', async () => {
    const { result, notify, gateway } = await setup();
    await act(() => result.current.handleUndoLastOrganization());
    expect(notify).toHaveBeenCalledWith(
      'restore',
      expect.stringContaining('1 arquivo(s) restaurado(s)')
    );
    expect(result.current.hasUndo).toBe(false);

    gateway.undoLastOrganization.mockRejectedValue(new SortlyError('NOTHING_TO_UNDO'));
    await act(() => result.current.handleUndoLastOrganization());
    expect(notify).toHaveBeenLastCalledWith('error', 'Nenhuma separação recente para desfazer.');
  });
});
