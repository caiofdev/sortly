import { describe, expect, it, vi } from 'vitest';
import { act, renderHook, waitFor } from '@testing-library/react';
import { SortlyError } from '../services/sortlyGateway';
import useSettings from './useSettings';

const view = (language, enabledKeys = ['byExtension']) => ({
  language,
  criteria: ['byDate', 'byExtension'].map((key) => ({
    key,
    enabled: enabledKeys.includes(key),
    locked: enabledKeys.length === 1 && enabledKeys.includes(key)
  }))
});

function fakeGateway(overrides = {}) {
  return {
    getSettings: vi.fn().mockResolvedValue(view('en')),
    setLanguage: vi.fn((language) => Promise.resolve(view(language))),
    setCriterion: vi.fn(() => Promise.resolve(view('en', ['byDate', 'byExtension']))),
    ...overrides
  };
}

describe('useSettings', () => {
  it('começa no padrão e carrega as preferências do backend', async () => {
    const gateway = fakeGateway();
    const { result } = renderHook(() => useSettings(gateway));
    expect(result.current.settings).toEqual({ language: 'pt-BR', criteria: [] });
    await waitFor(() => expect(result.current.settings).toEqual(view('en')));
  });

  it('falha ao carregar mantém o padrão', async () => {
    const gateway = fakeGateway({ getSettings: vi.fn().mockRejectedValue(new Error('x')) });
    const { result } = renderHook(() => useSettings(gateway));
    await waitFor(() => expect(gateway.getSettings).toHaveBeenCalled());
    expect(result.current.settings.language).toBe('pt-BR');
  });

  it('troca de idioma e de critério usam a visão devolvida pelo backend', async () => {
    const gateway = fakeGateway();
    const { result } = renderHook(() => useSettings(gateway));
    await waitFor(() => expect(result.current.settings.language).toBe('en'));

    await act(() => result.current.setLanguage('pt-BR'));
    expect(result.current.settings.language).toBe('pt-BR');

    await act(() => result.current.setCriterion('byDate', true));
    expect(gateway.setCriterion).toHaveBeenCalledWith('byDate', true);
    expect(result.current.settings.criteria.every((c) => c.enabled && !c.locked)).toBe(true);
  });

  it('erro do backend rejeita e mantém o estado anterior', async () => {
    const gateway = fakeGateway({
      setCriterion: vi.fn().mockRejectedValue(new SortlyError('LAST_CRITERION'))
    });
    const { result } = renderHook(() => useSettings(gateway));
    await waitFor(() => expect(result.current.settings.language).toBe('en'));

    await act(() =>
      expect(result.current.setCriterion('byExtension', false)).rejects.toMatchObject({
        code: 'LAST_CRITERION'
      })
    );
    expect(result.current.settings).toEqual(view('en'));
  });
});
