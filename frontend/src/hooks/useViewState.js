import { useCallback, useEffect, useMemo, useState } from 'react';
import * as Backend from '../../wailsjs/go/app/App';
import { EventsOn, OnFileDrop, OnFileDropOff } from '../../wailsjs/runtime/runtime';
import { DEFAULT_LANGUAGE } from '../i18n/language';

export const STATE_EVENT = 'sortly:state';

export const INITIAL_STATE = Object.freeze({
  version: 0,
  sourceFolderPath: '',
  destinationFolderPath: '',
  hasUndo: false,
  busy: '',
  settings: { language: DEFAULT_LANGUAGE, criteria: [] },
  notifications: []
});

// Evento e retorno do binding podem chegar fora de ordem: só um estado com
// versão maior substitui o atual, senão um "organizando" antigo travaria a tela.
export const newest = (current, next) => (next && next.version > current.version ? next : current);

// Fora da janela do Wails (npm run dev no navegador) não há runtime nem
// bindings: as assinaturas não fazem nada e as ações deixam o estado inicial.
const inWails = () => Boolean(window.runtime);

// Espelho do ViewState do backend (ADR 0005), único módulo que fala com o Go.
// Cada ação chama um binding e mostra o estado devolvido; o evento de estado
// cobre as mudanças no meio de uma ação (ex.: "organizando"). ready fica
// verdadeiro depois do primeiro GetState, para a tela não abrir no idioma
// padrão e trocar em seguida.
function useViewState() {
  const [state, setState] = useState(INITIAL_STATE);
  const [ready, setReady] = useState(false);

  const show = useCallback((next) => setState((current) => newest(current, next)), []);

  // Chamar dentro do then também captura o erro síncrono do binding
  // quando não há backend.
  const run = useCallback(
    (call) =>
      Promise.resolve()
        .then(() => call())
        .then(show)
        .catch(() => {}),
    [show]
  );

  useEffect(() => {
    run(Backend.GetState).finally(() => setReady(true));
  }, [run]);

  useEffect(() => (inWails() ? EventsOn(STATE_EVENT, show) : undefined), [show]);

  // O Wails entrega os caminhos só quando o drop termina num elemento com
  // --wails-drop-target: drop (useDropTarget = true).
  useEffect(() => {
    if (!inWails()) return undefined;
    OnFileDrop((_x, _y, paths) => run(() => Backend.DropPaths(paths)), true);
    return () => OnFileDropOff();
  }, [run]);

  const actions = useMemo(
    () => ({
      selectSource: () => run(Backend.SelectSource),
      selectDestination: () => run(Backend.SelectDestination),
      organize: () => run(Backend.Organize),
      undo: () => run(Backend.Undo),
      clearNotifications: () => run(Backend.ClearNotifications),
      setLanguage: (language) => run(() => Backend.SetLanguage(language)),
      setCriterion: (key, enabled) => run(() => Backend.SetCriterion(key, enabled))
    }),
    [run]
  );

  return { state, ready, actions };
}

export default useViewState;
