import { describe, expect, it, vi } from 'vitest';
import { fireEvent, render, screen } from '@testing-library/react';
import organizerCopy from '../i18n/organizerCopy';
import SettingsPage from './SettingsPage';

const pt = organizerCopy['pt-BR'];
const KEYS = ['byDuration', 'byPages', 'byResolution', 'byDate', 'bySize', 'byExtension'];

function renderSettings(criteria, theme = 'dark') {
  const onCriterionChange = vi.fn();
  const onThemeChange = vi.fn();
  render(
    <SettingsPage
      labels={pt}
      theme={theme}
      criteria={criteria}
      onThemeChange={onThemeChange}
      onCriterionChange={onCriterionChange}
    />
  );
  return Object.assign(onCriterionChange, { onThemeChange });
}

describe('SettingsPage', () => {
  it.each([
    ['dark', pt.themeDark],
    ['light', pt.themeLight]
  ])('tema %s: "%s" marcado no seletor de tema', (theme, pressed) => {
    renderSettings([], theme);
    const group = screen.getByRole('group', { name: pt.settingsTheme });
    expect(group.querySelector('[aria-pressed="true"]')).toHaveTextContent(pressed);
    expect(screen.getByText(pt.settingsHintTheme)).toBeInTheDocument();
  });

  it('escolher o tema claro pede a mudança ao backend', () => {
    const { onThemeChange } = renderSettings([]);
    fireEvent.click(screen.getByRole('button', { name: pt.themeLight }));
    expect(onThemeChange).toHaveBeenCalledWith('light');
  });

  it('um switch por critério, na ordem do backend', () => {
    renderSettings(KEYS.map((key) => ({ key, enabled: false, locked: false })));
    expect(screen.getAllByRole('switch').map((s) => s.getAttribute('aria-label'))).toEqual([
      'Duração (.mp4)',
      'Páginas',
      'Resolução',
      'Data',
      'Tamanho (MB)',
      'Extensão do arquivo'
    ]);
  });

  it.each([
    [false, pt.settingsHintByExtension, false],
    [true, pt.settingsLockedHint, true]
  ])('travado = %s: dica "%s" e switch desabilitado = %s', (locked, hint, disabled) => {
    renderSettings([{ key: 'byExtension', enabled: true, locked }]);
    expect(screen.getByText(hint)).toBeInTheDocument();
    expect(screen.getByRole('switch', { name: 'Extensão do arquivo' }).disabled).toBe(disabled);
  });

  it('ligar um critério pede a mudança ao backend', () => {
    const onCriterionChange = renderSettings([{ key: 'byDate', enabled: false, locked: false }]);
    fireEvent.click(screen.getByRole('switch', { name: 'Data' }));
    expect(onCriterionChange).toHaveBeenCalledWith('byDate', true);
  });
});
