import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, within } from '@testing-library/react';
import organizerCopy from '../i18n/organizerCopy';
import OrganizerView from './OrganizerView';

const pt = organizerCopy['pt-BR'];
const en = organizerCopy.en;

function renderView(props = {}) {
  const noop = vi.fn();
  return render(
    <OrganizerView
      language="pt-BR"
      criteria={[{ key: 'byExtension', enabled: true, locked: true }]}
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

const sidebar = () => screen.getByRole('navigation', { name: 'Sortly' });

describe('OrganizerView', () => {
  it('abre na página Organizar', () => {
    renderView();
    expect(screen.getByRole('heading', { name: pt.titleOrganize })).toBeInTheDocument();
    expect(screen.getByText(pt.dropTitle)).toBeInTheDocument();
  });

  it('a sidebar troca para Configurações e volta', () => {
    renderView();
    fireEvent.click(within(sidebar()).getByRole('button', { name: pt.navSettings }));
    expect(screen.getByRole('heading', { name: pt.titleSettings })).toBeInTheDocument();
    expect(screen.getByRole('switch', { name: pt.settingsByExtension })).toBeInTheDocument();
    fireEvent.click(within(sidebar()).getByRole('button', { name: pt.navOrganize }));
    expect(screen.getByRole('heading', { name: pt.titleOrganize })).toBeInTheDocument();
  });

  it('em inglês usa os textos em inglês', () => {
    renderView({ language: 'en' });
    expect(screen.getByRole('heading', { name: en.titleOrganize })).toBeInTheDocument();
  });

  it('o sino abre o painel; trocar de página o fecha', () => {
    renderView();
    fireEvent.click(screen.getByRole('button', { name: pt.notificationsTitle }));
    expect(screen.getByRole('dialog', { name: pt.notificationsTitle })).toBeInTheDocument();
    fireEvent.click(within(sidebar()).getByRole('button', { name: pt.navSettings }));
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });
});
