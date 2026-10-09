import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen, within } from '@testing-library/react';
import organizerCopy from '../i18n/organizerCopy';
import Sidebar from './Sidebar';

const labels = organizerCopy['pt-BR'];

function renderSidebar(page = 'organize') {
  const handlers = { onNavigate: vi.fn(), onLanguageChange: vi.fn() };
  render(<Sidebar labels={labels} page={page} language="pt-BR" {...handlers} />);
  return { nav: screen.getByRole('navigation', { name: 'Sortly' }), ...handlers };
}

describe('Sidebar', () => {
  it.each([
    ['organize', labels.navOrganize, labels.navSettings],
    ['history', labels.navHistory, labels.navOrganize],
    ['settings', labels.navSettings, labels.navOrganize]
  ])('página %s: "%s" é a atual e "%s" não', (page, current, other) => {
    const { nav } = renderSidebar(page);
    expect(within(nav).getByRole('button', { name: current })).toHaveAttribute(
      'aria-current',
      'page'
    );
    expect(within(nav).getByRole('button', { name: other })).not.toHaveAttribute('aria-current');
  });

  it('ordem do protótipo: Organizar, Histórico, Configurações', () => {
    const { nav } = renderSidebar();
    const items = [...nav.querySelectorAll('.st-nav')].map((b) => b.textContent);
    expect(items).toEqual([labels.navOrganize, labels.navHistory, labels.navSettings]);
  });

  it('clicar num item navega', () => {
    const { nav, onNavigate } = renderSidebar();
    fireEvent.click(within(nav).getByRole('button', { name: labels.navSettings }));
    expect(onNavigate).toHaveBeenCalledWith('settings');
  });

  it('logo decorativa, nome do app e idioma no rodapé', () => {
    const { nav, onLanguageChange } = renderSidebar();
    expect(nav.querySelector('img')).toHaveAttribute('alt', '');
    expect(within(nav).getByText('Sortly')).toBeInTheDocument();
    fireEvent.click(within(nav).getByRole('button', { name: 'EN' }));
    expect(onLanguageChange).toHaveBeenCalledWith('en');
  });
});
