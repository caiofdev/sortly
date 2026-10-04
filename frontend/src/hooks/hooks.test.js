import { describe, expect, it } from 'vitest';
import { act, renderHook } from '@testing-library/react';
import useNotifications, { MAX_NOTIFICATIONS } from './useNotifications';

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
