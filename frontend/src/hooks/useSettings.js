import { useCallback, useEffect, useState } from 'react';
import { DEFAULT_LANGUAGE } from '../i18n/language';
import defaultGateway from '../services/sortlyGateway';

const INITIAL_SETTINGS = { language: DEFAULT_LANGUAGE, criteria: [] };

// Preferências vindas do backend. As alterações devolvem a visão atualizada;
// em caso de erro a promessa é rejeitada e o estado continua o anterior.
function useSettings(gateway = defaultGateway) {
  const [settings, setSettings] = useState(INITIAL_SETTINGS);

  useEffect(() => {
    let isMounted = true;
    gateway
      .getSettings()
      .then((view) => {
        if (isMounted && view) setSettings(view);
      })
      .catch(() => {});
    return () => {
      isMounted = false;
    };
  }, [gateway]);

  const apply = useCallback(async (call) => setSettings(await call()), []);

  const setLanguage = useCallback(
    (language) => apply(() => gateway.setLanguage(language)),
    [apply, gateway]
  );
  const setCriterion = useCallback(
    (key, enabled) => apply(() => gateway.setCriterion(key, enabled)),
    [apply, gateway]
  );

  return { settings, setLanguage, setCriterion };
}

export default useSettings;
