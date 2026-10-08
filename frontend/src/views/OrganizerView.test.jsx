import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, within } from '@testing-library/react';
import organizerCopy from '../i18n/organizerCopy';
import OrganizerView from './OrganizerView';

const pt = organizerCopy['pt-BR'];
const en = organizerCopy.en;

function renderView(props = {}) {
  const noop = vi.fn();
  const onMarkNotificationsRead = vi.fn();
  const view = render(
    <OrganizerView
      language="pt-BR"
      criteria={[{ key: 'byExtension', enabled: true, locked: true }]}
      sourceFolderPath=""
      destinationFolderPath=""
      hasUndo={false}
      isLoading={false}
      loadingAction=""
      notifications={[]}
      unread={false}
      onClearNotifications={noop}
      onMarkNotificationsRead={onMarkNotificationsRead}
      onLanguageChange={noop}
      onCriterionChange={noop}
      onSelectSourceFolder={noop}
      onSelectDestinationFolder={noop}
      onOrganizeFiles={noop}
      onUndoLastOrganization={noop}
      {...props}
    />
  );
  return { ...view, onMarkNotificationsRead };
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

  it('abrir o painel marca como lidas; fechar pelo sino não marca de novo', () => {
    const { onMarkNotificationsRead } = renderView({ unread: true });
    const bell = screen.getByRole('button', { name: pt.notificationsTitle });
    fireEvent.click(bell);
    fireEvent.click(bell);
    expect(onMarkNotificationsRead).toHaveBeenCalledTimes(1);
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
  });

  it('"Marcar como lidas" marca e fecha o painel; Esc só fecha', () => {
    const { onMarkNotificationsRead } = renderView();
    const bell = screen.getByRole('button', { name: pt.notificationsTitle });
    fireEvent.click(bell);
    fireEvent.click(screen.getByRole('button', { name: pt.notificationsMarkRead }));
    expect(onMarkNotificationsRead).toHaveBeenCalledTimes(2);
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();

    fireEvent.click(bell);
    fireEvent.keyDown(screen.getByRole('dialog'), { key: 'Escape' });
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    expect(onMarkNotificationsRead).toHaveBeenCalledTimes(3);
  });

  it('cada notificação também aparece como toast', () => {
    renderView({
      notifications: [{ id: 1, kind: 'error', title: 'Pasta inválida.', text: '', time: '09:07' }]
    });
    expect(screen.getByRole('alert')).toHaveTextContent('Pasta inválida.');
  });
});
