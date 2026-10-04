import { describe, expect, it } from 'vitest';
import {
  DEFAULT_OPTIONS,
  OPTION_KEYS,
  canToggle,
  countSelected,
  normalizeOptions,
  toggleOption
} from './organizationOptions';

const all = Object.fromEntries(OPTION_KEYS.map((key) => [key, true]));

describe('organizationOptions', () => {
  it('mantém a ordem dos checkboxes', () => {
    expect(OPTION_KEYS).toEqual([
      'byDuration',
      'byPages',
      'byResolution',
      'byDate',
      'bySize',
      'byExtension'
    ]);
  });

  it('countSelected conta 0, 1 e 6 critérios', () => {
    expect(countSelected({})).toBe(0);
    expect(countSelected(DEFAULT_OPTIONS)).toBe(1);
    expect(countSelected(all)).toBe(6);
  });

  it.each([
    ['desmarcar o último (1 marcado) é bloqueado', DEFAULT_OPTIONS, 'byExtension', false, false],
    [
      'desmarcar com 2 marcados é permitido',
      { ...DEFAULT_OPTIONS, byDate: true },
      'byExtension',
      false,
      true
    ],
    ['marcar é sempre permitido', DEFAULT_OPTIONS, 'byPages', true, true],
    ['desmarcar algo já desmarcado é permitido', DEFAULT_OPTIONS, 'bySize', false, true]
  ])('canToggle: %s', (_, options, key, value, expected) => {
    expect(canToggle(options, key, value)).toBe(expected);
  });

  it('toggleOption aplica a mudança ou devolve o mesmo objeto', () => {
    expect(toggleOption(DEFAULT_OPTIONS, 'byDate', true)).toEqual({
      ...DEFAULT_OPTIONS,
      byDate: true
    });
    expect(toggleOption(DEFAULT_OPTIONS, 'byExtension', false)).toBe(DEFAULT_OPTIONS);
  });

  it.each([
    ['completa os padrões', { byDate: true }, { ...DEFAULT_OPTIONS, byDate: true }],
    [
      'byExtension false é respeitado',
      { byExtension: false, bySize: true },
      { ...DEFAULT_OPTIONS, byExtension: false, bySize: true }
    ],
    ['valor nulo vira o padrão', null, DEFAULT_OPTIONS]
  ])('normalizeOptions: %s', (_, saved, expected) => {
    expect(normalizeOptions(saved)).toEqual(expected);
  });
});
