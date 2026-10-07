import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import organizerCopy from '../i18n/organizerCopy';
import OrganizerView from './OrganizerView';

const pt = organizerCopy['pt-BR'];

function renderView(props = {}) {
  const noop = vi.fn();
  return render(
    <OrganizerView
      language="pt-BR"
      criteria={[]}
      sourceFolderPath=""
      destinationFolderPath=""
      hasUndo={false}
      isLoading={false}
      loadingAction=""
      notifications={[]}
      onClearNotifications={noop}
      onLanguageChange={noop}
      onCriterionChange={noop}
      onSelectSourceFolder={noop}
      onSelectDestinationFolder={noop}
      onOrganizeFiles={noop}
      onUndoLastOrganization={noop}
      {...props}
    />
  );
}

describe('OrganizerView', () => {
  it('sem pastas: textos de vazio e organizar desabilitado', () => {
    renderView();
    expect(screen.getByText(pt.sourceEmpty)).toBeInTheDocument();
    expect(screen.getByText(pt.destinationEmpty)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: pt.organize })).toBeDisabled();
  });

  it('com origem e destino: mostra os caminhos e habilita organizar', () => {
    renderView({ sourceFolderPath: 'C:\\origem', destinationFolderPath: 'C:\\destino' });
    expect(screen.getByText('C:\\origem')).toBeInTheDocument();
    expect(screen.getByText('C:\\destino')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: pt.organize })).toBeEnabled();
  });

  it('em inglês usa os textos em inglês', () => {
    renderView({ language: 'en' });
    expect(screen.getByRole('button', { name: organizerCopy.en.organize })).toBeInTheDocument();
  });

  it('o sino abre o painel de notificações e o fundo o fecha', () => {
    const { container } = renderView();
    const panel = () => container.querySelector('aside').className;
    expect(panel()).toContain('translate-x-full');
    fireEvent.click(screen.getByRole('button', { name: pt.notificationsTitle }));
    expect(panel()).toContain('translate-x-0');
    fireEvent.click(container.querySelector('aside').previousElementSibling);
    expect(panel()).toContain('translate-x-full');
  });
});
