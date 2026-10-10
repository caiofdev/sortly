import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import organizerCopy from '../i18n/organizerCopy';
import OrganizePage from './OrganizePage';

const pt = organizerCopy['pt-BR'];

function renderPage(props = {}) {
  const handlers = {
    onSelectSourceFolder: vi.fn(),
    onSelectDestinationFolder: vi.fn(),
    onOrganizeFiles: vi.fn(),
    onUndoLastOrganization: vi.fn()
  };
  render(
    <OrganizePage
      labels={pt}
      sourceFolderPath=""
      destinationFolderPath=""
      preview={{ status: '', totalFiles: 0, folders: [], otherFiles: 0 }}
      progress={{ done: 0, total: 0, file: '', folder: '' }}
      lastResult={null}
      hasUndo={false}
      isLoading={false}
      loadingAction=""
      {...handlers}
      {...props}
    />
  );
  return handlers;
}

describe('OrganizePage', () => {
  it('sem pastas: origem vazia, destino na própria origem e organizar desabilitado', () => {
    renderPage();
    expect(screen.getByText(pt.sourceEmpty)).toBeInTheDocument();
    expect(screen.getByText(pt.destinationEmpty)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: pt.organize })).toBeDisabled();
  });

  it.each([
    ['', pt.destinationSelect],
    ['C:\\destino', pt.destinationChange]
  ])('destino "%s": link "%s" escolhe o destino', (destination, action) => {
    const { onSelectDestinationFolder } = renderPage({ destinationFolderPath: destination });
    fireEvent.click(screen.getByRole('button', { name: action }));
    expect(onSelectDestinationFolder).toHaveBeenCalled();
  });

  it('com origem: mostra o caminho e habilita organizar', () => {
    const { onOrganizeFiles, onSelectSourceFolder } = renderPage({
      sourceFolderPath: 'C:\\origem'
    });
    expect(screen.getByText('C:\\origem')).toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: pt.organize }));
    fireEvent.click(screen.getByRole('button', { name: pt.dropSelectHintAction }));
    expect(onOrganizeFiles).toHaveBeenCalled();
    expect(onSelectSourceFolder).toHaveBeenCalled();
  });

  it('carregando outra ação: link do destino desabilitado', () => {
    renderPage({ isLoading: true, loadingAction: 'preview', destinationFolderPath: 'C:\\d' });
    expect(screen.getByRole('button', { name: pt.destinationChange })).toBeDisabled();
  });

  it('com resultado: o card mostra o Concluído no lugar da configuração', () => {
    const onStartOver = vi.fn();
    renderPage({
      hasUndo: true,
      lastResult: {
        movedFiles: 3,
        failedFiles: 0,
        destinationFolderPath: 'C:\\destino',
        folders: [{ name: 'pdf', count: 3 }],
        otherFiles: 0
      },
      onStartOver
    });
    expect(screen.getByText(pt.doneTitle)).toBeInTheDocument();
    expect(screen.queryByText(pt.dropTitle)).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: pt.startOver }));
    expect(onStartOver).toHaveBeenCalled();
  });

  it.each([
    ['organize', pt.organizing],
    ['restore', pt.restoring]
  ])('%s: o card mostra só o progresso ("%s") e o Cancelar', (loadingAction, title) => {
    const onCancel = vi.fn();
    renderPage({
      isLoading: true,
      loadingAction,
      sourceFolderPath: 'C:\\origem',
      progress: { done: 1, total: 4, file: 'a.pdf', folder: 'pdf' },
      onCancel
    });
    expect(screen.getByText(title)).toBeInTheDocument();
    expect(screen.getByRole('progressbar')).toHaveAttribute('aria-valuenow', '1');
    expect(screen.queryByText(pt.dropTitle)).not.toBeInTheDocument();
    fireEvent.click(screen.getByRole('button', { name: pt.cancel }));
    expect(onCancel).toHaveBeenCalled();
  });
});
