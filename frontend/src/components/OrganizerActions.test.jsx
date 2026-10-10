import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import organizerCopy from '../i18n/organizerCopy';
import OrganizerActions from './OrganizerActions';

const labels = organizerCopy['pt-BR'];

function renderActions(isLoading, { hasSource = true, hasUndo = true } = {}) {
  const handlers = { onOrganizeFiles: vi.fn(), onUndoLastOrganization: vi.fn() };
  render(
    <OrganizerActions
      labels={labels}
      isLoading={isLoading}
      hasUndo={hasUndo}
      hasSource={hasSource}
      {...handlers}
    />
  );
  return handlers;
}

describe('OrganizerActions', () => {
  it.each([
    ['tudo disponível', false, {}, false, false],
    ['ação em andamento desabilita os dois', true, {}, true, true],
    ['sem origem não organiza', false, { hasSource: false }, true, false],
    ['sem organização anterior não desfaz', false, { hasUndo: false }, false, true]
  ])('%s', (_name, isLoading, state, organizeDisabled, undoDisabled) => {
    renderActions(isLoading, state);
    expect(screen.getByRole('button', { name: labels.organize }).disabled).toBe(organizeDisabled);
    expect(screen.getByRole('button', { name: labels.undo }).disabled).toBe(undoDisabled);
  });

  it('primário amarelo para organizar, contorno vermelho para desfazer', () => {
    const { onOrganizeFiles, onUndoLastOrganization } = renderActions(false);
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
