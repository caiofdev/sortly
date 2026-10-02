// função                         | CC | casos
// estado inicial (lazy init)     |  4 | sem valor salvo; valor válido; byExtension ausente; JSON inválido
// updateOrganizationOption (set) |  4 | desmarcar o último (bloqueado); desmarcar com 2 marcados; marcar novo; desmarcar já desmarcado
//
// Valor-limite: quantidade de critérios marcados = 1 (bloqueia) e 2 (permite).

import { describe, expect, it } from 'vitest';
import { act, renderHook } from '@testing-library/react';
import useOrganizationOptions from './useOrganizationOptions';

const STORAGE_KEY = 'sortly.organizationOptions';

const defaults = {
  byDuration: false,
  byPages: false,
  byResolution: false,
  byDate: false,
  bySize: false,
  byExtension: true
};

function setup() {
  return renderHook(() => useOrganizationOptions());
}

describe('useOrganizationOptions — estado inicial', () => {
  it('usa os padrões quando não há nada salvo', () => {
    const { result } = setup();
    expect(result.current.organizationOptions).toEqual(defaults);
  });

  it('carrega o valor salvo, completando chaves ausentes', () => {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify({ byDate: true, byExtension: false }));

    const { result } = setup();

    expect(result.current.organizationOptions).toEqual({
      ...defaults,
      byDate: true,
      byExtension: false
    });
  });

  it('liga byExtension quando ele não está no valor salvo', () => {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify({ bySize: true }));

    const { result } = setup();

    expect(result.current.organizationOptions.byExtension).toBe(true);
    expect(result.current.organizationOptions.bySize).toBe(true);
  });

  it('volta aos padrões quando o valor salvo não é JSON válido', () => {
    window.localStorage.setItem(STORAGE_KEY, '{quebrado');

    const { result } = setup();

    expect(result.current.organizationOptions).toEqual(defaults);
  });
});

describe('useOrganizationOptions — alterar critério', () => {
  it('não permite desmarcar o último critério marcado (limite: 1)', () => {
    const { result } = setup();

    act(() => result.current.updateOrganizationOption('byExtension', false));

    expect(result.current.organizationOptions.byExtension).toBe(true);
  });

  it('permite desmarcar quando há dois critérios marcados (limite + 1)', () => {
    const { result } = setup();

    act(() => result.current.updateOrganizationOption('byDate', true));
    act(() => result.current.updateOrganizationOption('byExtension', false));

    expect(result.current.organizationOptions).toEqual({
      ...defaults,
      byDate: true,
      byExtension: false
    });
  });

  it('marca um critério novo', () => {
    const { result } = setup();

    act(() => result.current.updateOrganizationOption('byPages', true));

    expect(result.current.organizationOptions.byPages).toBe(true);
  });

  it('desmarcar um critério já desmarcado não altera nada', () => {
    const { result } = setup();

    act(() => result.current.updateOrganizationOption('bySize', false));

    expect(result.current.organizationOptions).toEqual(defaults);
  });

  it('salva as alterações no localStorage', () => {
    const { result } = setup();

    act(() => result.current.updateOrganizationOption('byResolution', true));

    expect(JSON.parse(window.localStorage.getItem(STORAGE_KEY))).toEqual({
      ...defaults,
      byResolution: true
    });
  });
});
