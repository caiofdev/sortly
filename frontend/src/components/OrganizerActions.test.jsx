import { describe, expect, it, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import organizerCopy from '../i18n/organizerCopy';
import OrganizerActions from './OrganizerActions';

const labels = organizerCopy['pt-BR'];

function renderActions(loadingAction) {
  return render(
    <OrganizerActions
      labels={labels}
      isLoading={loadingAction !== ''}
      loadingAction={loadingAction}
      hasUndo
      hasSource
      onOrganizeFiles={vi.fn()}
      onUndoLastOrganization={vi.fn()}
    />
  );
}

describe('OrganizerActions', () => {
  it.each([
    ['', labels.organize, 0],
    ['organize', labels.organizing, 1],
    ['restore', labels.organize, 1]
  ])('ação "%s": botão de organizar "%s" e %i spinner(s)', (action, organizeName, spinners) => {
    const { container } = renderActions(action);
    expect(screen.getByRole('button', { name: organizeName })).toBeInTheDocument();
    expect(container.querySelectorAll('svg.animate-spin')).toHaveLength(spinners);
  });

  it.each([
    ['organize', labels.organizing, labels.undo],
    ['restore', labels.undo, labels.organize]
  ])('em "%s", o botão da ação tem spinner e aria-busy', (action, busyName, idleName) => {
    renderActions(action);
    const busy = screen.getByRole('button', { name: busyName });
    expect(busy).toHaveAttribute('aria-busy', 'true');
    expect(busy.querySelector('svg.animate-spin')).not.toBeNull();
    expect(screen.getByRole('button', { name: idleName })).toHaveAttribute('aria-busy', 'false');
  });
});
