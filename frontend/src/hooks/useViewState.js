import { useCallback, useEffect, useMemo, useState } from 'react';
import { DEFAULT_LANGUAGE } from '../i18n/language';
import defaultGateway from '../services/sortlyGateway';

export const INITIAL_STATE = Object.freeze({
  sourceFolderPath: '',
  destinationFolderPath: '',
  hasUndo: false,
  busy: '',
  settings: { language: DEFAULT_LANGUAGE, criteria: [] },
  notifications: []
});

// Espelho do ViewState do backend. Cada ação chama um binding e mostra o
// estado devolvido; o evento de estado cobre as mudanças no meio de uma ação
// (ex.: "organizando"). ready fica verdadeiro depois do primeiro GetState,
// para a tela não abrir no idioma padrão e trocar em seguida.
function useViewState(gateway = defaultGateway) {
  const [state, setState] = useState(INITIAL_STATE);
  const [ready, setReady] = useState(false);

  const run = useCallback(
    (call) =>
      call()
        .then((next) => {
          if (next) setState(next);
        })
        .catch(() => {}),
    []
  );

  useEffect(() => {
    run(gateway.getState).finally(() => setReady(true));
  }, [gateway, run]);

  useEffect(() => gateway.subscribeState((next) => setState(next)), [gateway]);

  useEffect(
    () => gateway.subscribeFileDrop((paths) => run(() => gateway.dropPaths(paths))),
    [gateway, run]
  );

  const actions = useMemo(
    () => ({
      selectSource: () => run(gateway.selectSource),
      selectDestination: () => run(gateway.selectDestination),
      organize: () => run(gateway.organize),
      undo: () => run(gateway.undo),
      clearNotifications: () => run(gateway.clearNotifications),
      setLanguage: (language) => run(() => gateway.setLanguage(language)),
      setCriterion: (key, enabled) => run(() => gateway.setCriterion(key, enabled))
    }),
    [gateway, run]
  );

  return { state, ready, actions };
}

export default useViewState;
