// função                  | CC | casos
// usePersistentState      |  1 | lê o salvo; grava ao mudar
// readStorage/writeStorage|  2 | localStorage indisponível (leitura e gravação)
// useLanguagePreference   |  2 | idioma salvo válido; inválido
// useNotifications.notify |  1 | ordem (mais recente primeiro); limite 80/81; horário por idioma
// clearNotifications      |  1 | limpa
//
// Valor-limite: 80 notificações (todas ficam) e 81 (a mais antiga sai).

import { afterEach, describe, expect, it, vi } from 'vitest';
import { act, renderHook } from '@testing-library/react';
import useLanguagePreference from './useLanguagePreference';
import useNotifications, { MAX_NOTIFICATIONS } from './useNotifications';
import usePersistentState from './usePersistentState';

describe('usePersistentState', () => {
  afterEach(() => vi.restoreAllMocks());

  it('lê o valor salvo e grava ao mudar', () => {
    window.localStorage.setItem('k', '5');
    const { result } = renderHook(() =>
      usePersistentState('k', { deserialize: (v) => Number(v ?? 0), serialize: String })
    );
    expect(result.current[0]).toBe(5);
    act(() => result.current[1](6));
    expect(window.localStorage.getItem('k')).toBe('6');
  });

  it('localStorage indisponível não quebra a interface', () => {
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('bloqueado');
    });
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('cota');
    });
    const { result } = renderHook(() =>
      usePersistentState('k', { deserialize: (v) => v ?? 'padrão' })
    );
    expect(result.current[0]).toBe('padrão');
    expect(() => act(() => result.current[1]('novo'))).not.toThrow();
  });
});

describe('useLanguagePreference', () => {
  it.each([
    ['en', 'en'],
    ['fr', 'pt-BR'],
    [null, 'pt-BR']
  ])('salvo %s → %s (formato texto puro, compatível)', (saved, expected) => {
    if (saved) window.localStorage.setItem('sortly.language', saved);
    const { result } = renderHook(() => useLanguagePreference());
    expect(result.current.language).toBe(expected);
    expect(window.localStorage.getItem('sortly.language')).toBe(expected);
  });
});

describe('useNotifications', () => {
  it('mais recente primeiro, com id, tipo, mensagem e horário', () => {
    const { result } = renderHook(() => useNotifications('pt-BR'));
    act(() => result.current.notify('info', 'primeira'));
    act(() => result.current.notify('error', 'segunda'));
    const [latest, older] = result.current.notifications;
    expect(latest).toMatchObject({ type: 'error', message: 'segunda' });
    expect(older).toMatchObject({ type: 'info', message: 'primeira' });
    expect(latest.id).not.toBe(older.id);
    expect(latest.time).toMatch(/^\d{2}:\d{2}$/);
  });

  it('80 ficam; a 81ª remove a mais antiga', () => {
    const { result } = renderHook(() => useNotifications('pt-BR'));
    act(() => {
      for (let i = 1; i <= MAX_NOTIFICATIONS; i += 1) result.current.notify('info', `n${i}`);
    });
    expect(result.current.notifications).toHaveLength(80);
    expect(result.current.notifications.at(-1).message).toBe('n1');

    act(() => result.current.notify('info', 'n81'));
    expect(result.current.notifications).toHaveLength(80);
    expect(result.current.notifications[0].message).toBe('n81');
    expect(result.current.notifications.at(-1).message).toBe('n2');
  });

  it('horário no formato do idioma atual e limpar', () => {
    const { result, rerender } = renderHook(({ lang }) => useNotifications(lang), {
      initialProps: { lang: 'pt-BR' }
    });
    rerender({ lang: 'en' });
    act(() => result.current.notify('info', 'x'));
    expect(result.current.notifications[0].time).toMatch(/^\d{2}:\d{2} [AP]M$/);
    act(() => result.current.clearNotifications());
    expect(result.current.notifications).toEqual([]);
  });
});
