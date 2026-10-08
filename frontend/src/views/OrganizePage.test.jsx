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

  it('organizando: link do destino desabilitado', () => {
    renderPage({ isLoading: true, loadingAction: 'organize', destinationFolderPath: 'C:\\d' });
    expect(screen.getByRole('button', { name: pt.destinationChange })).toBeDisabled();
  });
});
