import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import organizerCopy from '../i18n/organizerCopy';
import OrganizerActions from './OrganizerActions';

const labels = organizerCopy['pt-BR'];

function renderActions(loadingAction, { hasSource = true, hasUndo = true } = {}) {
  const handlers = { onOrganizeFiles: vi.fn(), onUndoLastOrganization: vi.fn() };
  render(
    <OrganizerActions
      labels={labels}
      isLoading={loadingAction !== ''}
      loadingAction={loadingAction}
      hasUndo={hasUndo}
      hasSource={hasSource}
      {...handlers}
    />
  );
  return handlers;
}

describe('OrganizerActions', () => {
  it.each([
    ['', labels.organize, labels.undo],
    ['organize', labels.organizing, labels.undo],
    ['restore', labels.organize, labels.restoring]
  ])('ação "%s": botões "%s" e "%s"', (action, organizeName, undoName) => {
    renderActions(action);
    expect(screen.getByRole('button', { name: organizeName })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: undoName })).toBeInTheDocument();
  });

  it.each([
    ['organize', labels.organizing, labels.undo],
    ['restore', labels.restoring, labels.organize]
  ])('em "%s", só o botão da ação gira e tem aria-busy', (action, busyName, idleName) => {
    renderActions(action);
    const busy = screen.getByRole('button', { name: busyName });
    const idle = screen.getByRole('button', { name: idleName });
    expect(busy).toHaveAttribute('aria-busy', 'true');
    expect(busy).toHaveClass('st-btn--busy');
    expect(idle).toHaveAttribute('aria-busy', 'false');
    expect(idle).not.toHaveClass('st-btn--busy');
  });

  it.each([
    ['tudo disponível', '', {}, false, false],
    ['ação em andamento desabilita os dois', 'organize', {}, true, true],
    ['sem origem não organiza', '', { hasSource: false }, true, false],
    ['sem organização anterior não desfaz', '', { hasUndo: false }, false, true]
  ])('%s', (_name, action, state, organizeDisabled, undoDisabled) => {
    renderActions(action, state);
    const organizeName = action === 'organize' ? labels.organizing : labels.organize;
    expect(screen.getByRole('button', { name: organizeName }).disabled).toBe(organizeDisabled);
    expect(screen.getByRole('button', { name: labels.undo }).disabled).toBe(undoDisabled);
  });

  it('primário amarelo para organizar, contorno vermelho para desfazer', () => {
    const { onOrganizeFiles, onUndoLastOrganization } = renderActions('');
    const organize = screen.getByRole('button', { name: labels.organize });
    const undo = screen.getByRole('button', { name: labels.undo });
    expect(organize).toHaveClass('st-btn--primary');
    expect(undo).toHaveClass('st-btn--danger');
    fireEvent.click(organize);
    fireEvent.click(undo);
    expect(onOrganizeFiles).toHaveBeenCalled();
    expect(onUndoLastOrganization).toHaveBeenCalled();
  });
});
